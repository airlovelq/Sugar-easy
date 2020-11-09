import logging
import os
logging.basicConfig(level=logging.INFO, 
                    format='%(asctime)s %(name)s %(levelname)s %(message)s', 
                    filename='/root/logs/{}.log'.format(os.environ.get('WORKER_NAME', 'worker')))
logger = logging.getLogger('worker')
