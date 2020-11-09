import os
from importlib import import_module
import logging

logger = logging.Logger(name=__name__)

class PredictWorker(object):
    def __init__(self):
        self._model_inst = None
        self._get_model_inst()
        
    def _get_model_class(self):
        model_class = os.environ.get('MODEL_CLASS','Model')
        model_file = os.environ.get('MODEL_FILE', 'model')
        try:
            # Import model file as module
            mod = import_module('model.'+model_file)
            # Extract model class from module
            clazz = getattr(mod, model_class)
        except Exception as e:
            logger.log(logging.INFO, 'Getting Model Class Failed')
            logger.log(logging.ERROR, repr(e))
            raise e
        finally:
            pass
        return clazz

    def _get_model_inst(self):
        clazz = self._get_model_class()
        self._model_inst = clazz()
        self._model_inst.load('/root/params/'+os.environ.get('MODEL_PARAM_FILE', 'param'))

    def predict(self, query):
        result = self._model_inst.predict(query)
        return result