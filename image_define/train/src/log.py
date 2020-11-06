import logging

logging.basicConfig(level=logging.INFO, 
                    format='%(asctime)s %(name)s %(levelname)s %(message)s', 
                    filename='/root/train/logs/train_worker.log')
logger = logging.getLogger('train_worker')
