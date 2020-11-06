#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#
import numpy as np
from time import time
import os
import base64
import pickle
import argparse
import lightgbm as lgb
from sklearn.model_selection import train_test_split
from sklearn.metrics import log_loss,accuracy_score
import logging
import msgpack
import io
from PIL import Image
import csv
import tempfile
import zipfile
from utils import load_table_classification_dataset
import pandas as pd
logger = logging.getLogger(__name__)

class LightGBM(object):
    '''
    Implements LeNet5 network to train image classification model on mnist dataset
    '''
    @staticmethod
    def get_param_type():
        return {'num_leaves': 'int', 'max_depth': 'int', 'learning_rate': 'float'}

    def __init__(self, **knobs):
        # super().__init__(**knobs)
        self._model = None
        self._feature_list = knobs.get('feature_list')
        self._target = knobs.get('target')
        self._num_classes = knobs.get('num_classes', 10)

    def train(self, dataset_path, **train_args):
        learning_rate = float(train_args['learning_rate'])
        num_leaves = int(train_args['num_leaves'])
        max_depth = int(train_args['max_depth'])
        boosting_type = train_args.get('boosting_type', 'gbdt')

        features, classes, num_samples, num_classes = load_table_classification_dataset(dataset_path, self._feature_list, self._target)
        self._num_classes = num_classes
        
        train = {}
        train['features'] = features
        train['classes'] = classes
        validation = {}
        train['features'], validation['features'], train['classes'], validation[
            'classes'] = train_test_split(train['features'],
                                          train['classes'],
                                          test_size=0.2,
                                          random_state=0)
        # X_train, y_train = features, classes
        X_train, y_train = train['features'], train['classes']
                                                
        X_validation, y_validation = validation['features'],validation['classes']
        train_set = lgb.Dataset(X_train, y_train)
        validate_set = lgb.Dataset(X_validation, y_validation)
        
        lgb_params = {'task': 'train',
                    'boosting_type': boosting_type,
                    'objective': 'multiclass',
                    'num_class': self._num_classes,
                    'metric': 'multi_logloss',
                    'learning_rate': learning_rate,
                    'max_depth': max_depth,
                    'num_leaves': num_leaves}
        lgb_params = {**lgb_params, **train_args}

        abc = {}
        self._model = lgb.train(lgb_params, train_set, valid_sets=[train_set, validate_set], callbacks=[lgb.record_evaluation(abc)])
    
        # Compute train accuracy
        train_loss = abc['training']['multi_logloss'][-1]
        
        logger.info('Train loss: {}'.format(train_loss))
        # logger.info('Train accuracy: {}'.format(train_acc))

    def evaluate(self, dataset_path):
        features, classes, num_samples, num_classes = load_table_classification_dataset(dataset_path, self._feature_list, self._target)
        classes_pred = self._model.predict(features)
        # train_set = lgb.Dataset(features, classes)
        # abc = {}
        # self._model.train_set = self._train_set
        # result=self._model.eval(data=train_set, name='train')
        loss = log_loss(classes, classes_pred)
        argm = np.argmax(classes_pred, axis=1)
        acc = accuracy_score(classes, np.argmax(classes_pred, axis=1))
        return loss, acc

    def predict(self, query):
        pquery = [query.decode().split(',')]
        # pquery = pd.DataFrame(data=l, columns=self._feature_list)
        result = self._model.predict(pquery)
        res = np.argmax(result[0])
        return res

    def destroy(self):
        pass

    def save(self, path):
        param_path = os.path.join(path, 'param')
        self._model.save_model(param_path)

        params = {}
        config_path = os.path.join(path, 'config')
        params['feature_list'] = self._feature_list
        params['target'] = self._target
        params['num_classes'] = self._num_classes
        params_bytes = msgpack.packb(params, use_bin_type=True)
        with open(config_path, 'wb') as f:
            f.write(params_bytes)


    def load(self, path):
        param_path = os.path.join(path, 'param')
        self._model = lgb.Booster(model_file=param_path)
   
        config_path = os.path.join(path, 'config')
        with open(config_path, 'rb') as f:
            params_bytes = f.read()
            params = msgpack.unpackb(params_bytes, raw=False)
            self._feature_list = params['feature_list']
            self._target = params['target']
            self._num_classes = int(params['num_classes'])


# _model = 
# _model.

# params={
#                      'boosting_type': 'gbdt',
#                 'objective': 'binary',
#                'metric': 'cross_entropy',
#                 'nthread':4,
#                  'n_estimators':10,
#             'learning_rate':0.02,
#             'num_leaves':self._knobs['num_leaves'],
#             'colsample_bytree':self._knobs['colsample_bytree'],
#             'subsample':self._knobs['subsample'],
#             'max_depth':self._knobs['max_depth'],
#             # 'reg_alpha':0.041545473,
#             # 'reg_lambda':0.0735294,
#             # 'min_split_gain':0.0222415,
#             # 'min_child_weight':39.3259775,
          
#             'verbose':-1
             
#             }