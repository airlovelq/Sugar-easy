import os
import json
# from utils import load_model_class
# from log import logger
from importlib import import_module
import logging
from hyperopt import fmin, tpe, hp, partial

logger = logging.getLogger(__name__)

class ModelType(object):
    IMAGE_CLASSIFICATION = 'IMAGE_CLASSIFICATION'
    TABLE_CLASSIFICATION = 'TABLE_CLASSIFICATION'

def load_model_class(module, model_class):
    clazz = None
    try:
        # Import model file as module
        mod = import_module(module)
        # Extract model class from module
        clazz = getattr(mod, model_class)
    except Exception as e:
        raise e
        # logger.log(logging.INFO, 'Getting Model Class Failed')
        # logger.log(logging.ERROR, repr(e))
        # raise e
    finally:
        pass
    return clazz

class TrainWorker(object):
    def __init__(self, **params):
        self._model_inst = None
        self._train_dataset = None
        self._validate_dataset = None
        self._train_params = {}
        self._model_params = {}
        self._model_ckpt = None
        self._save_param_path = None
        self._use_automl = False
        self._advisor = {}
        self._best_loss = None
        self._trial_count = 0
        self._dest_score = None
        self._model_type = ModelType.IMAGE_CLASSIFICATION
        # self._save_checkpoint_path = None
    
    def _get_advisor(self):
        logger.log(logging.INFO, 'Start Getting Train job Advisor')
        try:
            param_types = self._model_class.get_param_type()
            for param_key, param_value_list in self._train_params.items():
                if param_value_list[0] == 'choice':
                    self._advisor[param_key] = hp.choice(param_key, param_value_list[1])
                else:
                    if param_types.get(param_key) == 'int':
                        self._advisor[param_key] = hp.uniformint(param_key, param_value_list[1][0], param_value_list[1][1])
                    else:
                        self._advisor[param_key] = hp.uniform(param_key, param_value_list[1][0], param_value_list[1][1])
            logger.log(logging.INFO, 'Finish Getting Train job Advisor')
        except Exception as e:
            logger.log(logging.ERROR, repr(e))
    
    def _train_once(self, **params):
        logger.log(logging.INFO, 'Train Params is {}'.format(params))
        model_inst = self._model_class(**self._model_params)
        if self._model_ckpt is not None:
            logger.log(logging.INFO, 'Loading Checkpoint')
            model_inst.load(self._model_ckpt)
        self._trial_count += 1
        logger.log(logging.INFO, 'Start Training Trial {}, param is {}'.format(self._trial_count, params))
        model_inst.train(self._train_dataset, **params)
        logger.log(logging.INFO, 'Training Finished')
        logger.log(logging.INFO, 'Start Evaluating')
        res_eval = model_inst.evaluate(self._validate_dataset)
        logger.log(logging.INFO, 'Evaluating Finished')
        logger.log(logging.INFO, 'Score is {}'.format(res_eval))
        loss = res_eval[0]
        if self._best_loss is None or self._best_loss > loss:
            logger.log(logging.INFO, 'Start Saving')
            model_inst.save(self._save_param_path)
            logger.log(logging.INFO, 'Save Finished ')
            self._best_loss = loss
        model_inst.destroy()
        return loss

    def train(self):
        logger.log(logging.INFO, 'Loading Dataset')
        self._load_dataset()
        logger.log(logging.INFO, 'Loading Model')
        self._load_model()
        logger.log(logging.INFO, 'Loading Model Params')
        self._load_model_params()
        logger.log(logging.INFO, 'Loading Train Params')
        self._load_train_params()
        logger.log(logging.INFO, 'Loading Checkpoint Path')
        self._load_model_checkpoint()
        logger.log(logging.INFO, 'Loading Save Path')
        self._load_save_path()
        logger.log(logging.INFO, 'Loading Use Automl')
        self._load_use_automl()
        logger.log(logging.INFO, 'Loading Max Trials')
        self._load_max_trials()
        logger.log(logging.INFO, 'Loading Dest Score')
        self._load_dest_score()
        # logger.log(logging.INFO, 'Loading Checkpoint')
        # if self._model_ckpt is not None:
        #     self._model_inst.load(self._model_ckpt)
        logger.log(logging.INFO, 'Training...')
        if self._use_automl:
            self._get_advisor()
            fn = lambda params : self._train_once(**params)
            algo = partial(tpe.suggest, n_startup_jobs=1)
            fmin(fn, self._advisor, algo, max_evals=self._max_trials, pass_expr_memo_ctrl=None)
        else:
            self._train_once(**self._train_params)
            
        # logger.log(logging.INFO, 'Destroy Model')
        # self._model_inst.destroy()

    def _load_model_params(self):
        params_str = os.environ.get('MODEL_PARAMS', '{}')
        self._model_params = json.loads(params_str)

    def _load_train_params(self):
        params_str = os.environ.get('TRAIN_PARAMS', '{}')
        self._train_params = json.loads(params_str)

    def _load_dataset(self):
        self._train_dataset = './dataset/'+os.environ.get('TRAIN_DATASET', None)
        self._validate_dataset = './dataset/'+os.environ.get('VALIDATE_DATASET', None)

    def _load_model(self):
        model_class = load_model_class('model.'+os.environ.get('MODEL_FILE', 'model'), os.environ.get('MODEL_CLASS','Model'))
        self._model_class = model_class

    def _load_model_checkpoint(self):
        self._model_ckpt = None if os.environ.get('MODEL_CHECKPOINT', None) is None else './params/' + os.environ.get('MODEL_CHECKPOINT', None)

    def _load_save_path(self):
        self._save_param_path = './params/'+os.environ.get('MODEL_SAVE_PATH')

    def _load_use_automl(self):
        self._use_automl = int(os.environ.get('USE_AUTOML', 0))
    
    def _load_dest_score(self):
        self._dest_score = float(os.environ.get('DEST_SCORE', 0))

    def _load_max_trials(self):
        self._max_trials = int(os.environ.get('MAX_TRIALS'))
   