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
# from model import BaseModel

import numpy as np
from time import time
import os
import base64
import pickle
import argparse
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import models, layers

from tensorflow.keras.layers import Dense, Flatten, Conv2D, AveragePooling2D
from tensorflow.keras.optimizers import Adam
from tensorflow.keras.models import Sequential
from tensorflow.keras.preprocessing.image import ImageDataGenerator
from tensorflow.keras.utils import to_categorical
from tensorflow.keras.callbacks import TensorBoard
from sklearn.model_selection import train_test_split
import logging
import msgpack
import io
from PIL import Image
import csv
import tempfile
import zipfile
from utils import load_image_classification_dataset

logger = logging.getLogger(__name__)

class Resnet(object):
    '''
    Implements LeNet5 network to train image classification model on mnist dataset
    '''
    @staticmethod
    def get_param_type():
        return {'batch_size': 'int', 'max_epochs': 'int', 'learning_rate': 'float'}

    def __init__(self, **knobs):
        # super().__init__(**knobs)
        self._model = None
        self._image_size = knobs.get('image_size', 32)
        # self._num_classes = knobs.get('num_classes', 10)

    def train(self, dataset_path, **train_args):
        batch_size = train_args.get('batch_size')
        max_epochs = train_args.get('max_epochs')
        learning_rate = float(train_args['learning_rate'])
        
        pil_images, image_classes, num_samples, num_classes = load_image_classification_dataset(dataset_path, 'RGB')
        self._num_classes = num_classes
        images = self._trans_image(pil_images)
        # pil_images = [x.resize([self._image_size, self._image_size]) for x in pil_images]

        # # Convert to numpy arrays
        # images = [np.asarray(x) for x in pil_images]
        # images = np.asarray(images) / 255

        # images = self._prepare_X(images)
        train = {}
        train['images'] = images
        train['classes'] = image_classes
        validation = {}
        train['images'], validation['images'], train['classes'], validation[
            'classes'] = train_test_split(train['images'],
                                          train['classes'],
                                          test_size=0.2,
                                          random_state=0)
        # train['images'] = np.pad(train['images'],
        #                          ((0, 0), (2, 2), (2, 2), (0, 0)), 'constant')
        # validation['images'] = np.pad(validation['images'],
        #                               ((0, 0), (2, 2), (2, 2), (0, 0)),
        #                               'constant')

        X_train, y_train = train['images'], to_categorical(train['classes'],
                                                           num_classes=num_classes)
        X_validation, y_validation = validation['images'], to_categorical(
            validation['classes'], num_classes=self._num_classes)

        train_generator = ImageDataGenerator().flow(X_train,
                                                    y_train,
                                                    batch_size=batch_size)
        validation_generator = ImageDataGenerator().flow(X_validation,
                                                         y_validation,
                                                         batch_size=batch_size)
        steps_per_epoch = X_train.shape[0] // batch_size
        validation_steps = X_validation.shape[0] // batch_size

        tensorboard = TensorBoard(log_dir="logs/{}".format(time()))
        if not self._model:
            self._model = self._build_classifier(float(learning_rate))
        logger.log(logging.INFO, 'Begin loading train dataset')
        self._model.fit_generator(train_generator,
                                  steps_per_epoch=steps_per_epoch,
                                  epochs=max_epochs,
                                  validation_data=validation_generator,
                                  validation_steps=validation_steps,
                                  shuffle=True,
                                  callbacks=[tensorboard])
    
        # Compute train accuracy
        (train_loss, train_acc) = self._model.evaluate(X_validation,
                                                       y_validation)
        logger.info('Train loss: {}'.format(train_loss))
        logger.info('Train accuracy: {}'.format(train_acc))

    def evaluate(self, dataset_path):
        pil_images, image_classes, num_samples, num_classes = load_image_classification_dataset(dataset_path, 'RGB')
        images = self._trans_image(pil_images)
        # images = np.pad(images, ((0, 0), (2, 2), (2, 2), (0, 0)), 'constant')
        X_test, y_test = images, to_categorical(image_classes, num_classes=self._num_classes)

        # Compute test accuracy
        (test_loss, test_acc) = self._model.evaluate(X_test, y_test)
        return test_loss, test_acc

    def predict(self, query):
        encoded = io.BytesIO(query)
        pil_image = Image.open(encoded).convert('RGB')
        images = self._trans_image([pil_image])
        # images = [np.asarray(pil_image.resize([self._image_size, self._image_size])) / 255] 
        # X = self._prepare_X(images)
        # X = np.pad(X, ((0, 0), (2, 2), (2, 2), (0, 0)), 'constant')
        probs = self._model.predict(images)

        lres = np.argmax(probs[0])
        res = str(lres)
        return res

    def destroy(self):
        pass

    def save(self, path):
        self._model.save(path)
        params = {}
        # Put model parameters
        # model_bytes = pickle.dumps(self._model)
        # model_base64 = base64.b64encode(model_bytes).decode('utf-8')
        # params['model_base64'] = model_base64

        # Put image size
        config_path = os.path.join(path, 'config')
        params['image_size'] = self._image_size
        params['num_classes'] = self._num_classes
        params_bytes = msgpack.packb(params, use_bin_type=True)
        with open(config_path, 'wb') as f:
            f.write(params_bytes)

    def load(self, path):
        self._model = tf.keras.models.load_model(path)
            
        config_path = os.path.join(path, 'config')
        with open(config_path, 'rb') as f:
            params_bytes = f.read()
            params = msgpack.unpackb(params_bytes, raw=False)
            self._image_size = int(params['image_size'])
            self._num_classes = int(params['num_classes'])

    def _trans_image(self, pil_images):
        pil_images = [x.resize([self._image_size, self._image_size]) for x in pil_images]

        # Convert to numpy arrays
        images = [np.asarray(x) for x in pil_images]
        images = np.asarray(images)
        return images

    def _build_classifier(self, l_rate):
        model = tf.keras.applications.ResNet50(input_shape=(self._image_size, self._image_size, 3),weights=None,classes=self._num_classes,include_top=True)
        adam = Adam(lr=l_rate,
                    beta_1=0.9,
                    beta_2=0.999,
                    epsilon=None,
                    decay=0.0,
                    amsgrad=False)
        model.compile(loss=tf.keras.losses.categorical_crossentropy,
                      optimizer=adam,
                      metrics=['accuracy'])
        return model