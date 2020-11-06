import os
os.environ['MODEL_PARAM_FILE'] = 'save'
os.environ['MODEL_FILE'] = 'vgg16'
os.environ['MODEL_CLASS'] = 'Vgg16'

from flask import Flask, request, g, jsonify, make_response, Response
from flask_cors import CORS
import pickle

import os
import threading
from image_define.predict.src.predictor import PredictWorker

install_command = 'pip3 install -r ./model/requirements.txt'
exit_code = os.system(install_command)
if exit_code != 0: 
    raise Exception('Install command gave non-zero exit code: "{}"'.format(install_command))

c = 0
predictor = PredictWorker()
mutex = threading.Lock()
app = Flask(__name__)
CORS(app)


@app.errorhandler(Exception)
def exception_handle(ex):
    return repr(ex), 500

@app.route('/app/predict', methods=['POST'])
def predict():
    mutex.acquire(timeout=60)
    global c
    c+=1
    print(c)
    result = predictor.predict(request.data)
    mutex.release()
    # res = Response(result, content_type='application/octet-stream')
    # res.headers["Content-disposition"] = 'attachment; filename={}'.format('app')
    # return res
    return result

if __name__ == '__main__':
    # url = worker.get_url()
    # if url is not None:
    print(c)
    app.config['JSON_AS_ASCII'] = False
    app.run(host='0.0.0.0',
            port=int(os.environ.get('APP_PORT', 7000)),
            debug=False,
            threaded=True)