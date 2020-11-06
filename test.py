# from lenet5 import LeNet5
# from vgg16 import Vgg16
# from vgg19 import Vgg19
# from resnet import Resnet
# from inceptionresnetv2 import InceptionResNetV2
# from inceptionv3 import InceptionV3
# from densenet201 import DenseNet201
from ligthgbm_model import LightGBM
import os
# import tensorflow as tf
# os.environ["CUDA_VISIBLE_DEVICES"] = "0"
# gpus = tf.config.experimental.list_physical_devices(device_type='GPU')
# for gpu in gpus:
#     tf.config.experimental.set_memory_growth(gpu, True)
model = LightGBM(feature_list=['a','b','c'], target='d')

model.train('./test.csv', num_leaves=2, max_depth=2, learning_rate=0.1)
res = model.predict('1,4,5'.encode())
print(res)
res = model.evaluate('./test.csv')
model.save('/home/luqin/code/models/test/save')

# model.load('/home/luqin/code/models/test/save')
# with open('/home/luqin/fashion_mnist_train/1-31832.png', 'rb') as f:
#     print(model.predict(f.read()))