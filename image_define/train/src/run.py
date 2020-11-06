from worker import TrainWorker
import os

install_command = 'pip3 install -r /root/model/requirements.txt'
exit_code = os.system(install_command)
if exit_code != 0: 
    raise Exception('Install command gave non-zero exit code: "{}"'.format(install_command))

if __name__ == "__main__":
    worker = TrainWorker()
    worker.train()