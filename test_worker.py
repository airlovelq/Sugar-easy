from image_define.train.src.worker import TrainWorker
import os
import json
os.environ['MODEL_FILE'] = 'vgg16'
os.environ['MODEL_CLASS'] = 'Vgg16'
os.environ['USE_AUTOML'] = '1'
os.environ['DEST_SCORE'] = '1'
model_params = {'image_size': 32}
os.environ['MODEL_PARAMS'] = json.dumps(model_params)
train_params = {'batch_size': ('range', (16, 48)), 'max_epochs': ('choice', (2,3,4,6)), 'learning_rate': ('range', (0.0001, 0.001))}
os.environ['TRAIN_PARAMS'] = json.dumps(train_params)
os.environ['MODEL_CHECKPOINT'] = './params/load'
os.environ['TRAIN_DATASET'] = 'fashion_mnist_train'
os.environ['VALIDATE_DATASET'] = 'fashion_mnist_train'
os.environ['MODEL_SAVE_PATH'] = 'save'
os.environ['MAX_TRIALS'] = '3'
worker = TrainWorker()
worker.train()