from worker import TrainWorker
import os

install_command = 'pip3 install -r ./model/requirements.txt'
try_count = 0
while True:
    exit_code = os.system(install_command)
    if exit_code == 0:
        break
    else:
        try_count += 1
        if try_count == 10: 
            raise Exception('Install command gave non-zero exit code: "{}"'.format(install_command))

if __name__ == "__main__":
    worker = TrainWorker()
    worker.train()