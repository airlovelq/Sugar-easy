import os
# from importlib import import_module
from utils import load_model_class
import logging

logger = logging.Logger(name=__name__)

class PredictWorker(object):
    def __init__(self):
        self._model_inst = None
        self._get_model_inst()

    def _get_model_inst(self):
        clazz = load_model_class('model.'+os.environ.get('MODEL_FOLDER', 'model')+'.'+os.environ.get('MODEL_FILE', 'model'), os.environ.get('MODEL_CLASS','Model'))
        self._model_inst = clazz()
        self._model_inst.load('./params/'+os.environ.get('MODEL_PARAM', 'param'))

    def predict(self, query):
        result = self._model_inst.predict(query)
        return result