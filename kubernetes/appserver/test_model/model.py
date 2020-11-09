import time

class Model(object):
    def __init__(self):
        pass

    def load(self, param_path):
        print('loading param')

    def predict(self, query):
        for i in range(10000000):
            i = i+2
            i = i*3
            for j in range(1000000):
                i = i + j
        return str(time.localtime(time.time()))