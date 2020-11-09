from flask import Flask, request, g, jsonify, make_response, Response
from flask_cors import CORS
import pickle

import os
from predictor import PredictWorker

install_command = 'pip3 install -r ./model/requirement.txt'
exit_code = os.system(install_command)
if exit_code != 0: 
    raise Exception('Install command gave non-zero exit code: "{}"'.format(install_command))

predictor = PredictWorker()

app = Flask(__name__)
CORS(app)

@app.errorhandler(Exception)
def exception_handle(ex):
    return repr(ex), 500

@app.route('/app/predict', methods=['POST'])
def predict():
    result = predictor.predict(request.data)
    # res = Response(result, content_type='application/octet-stream')
    # res.headers["Content-disposition"] = 'attachment; filename={}'.format('app')
    # return res
    return result

if __name__ == '__main__':
    # url = worker.get_url()
    # if url is not None:
    app.config['JSON_AS_ASCII'] = False
    app.run(host='0.0.0.0',
            port=int(os.environ.get('APP_PORT', 7000)),
            debug=True,
            threaded=True)