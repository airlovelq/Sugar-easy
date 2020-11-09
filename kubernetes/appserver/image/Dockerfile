FROM ubuntu:18.04

RUN apt-get update

RUN apt-get install -y python3
RUN apt-get install -y python3-pip

WORKDIR /root

RUN mkdir /root/.pip
COPY image/pip.conf /root/.pip

COPY image/requirements.txt /root/requirements.txt
RUN pip3 install -r /root/requirements.txt

RUN mkdir /root/model

COPY image/run_app.py /root/
COPY image/predictor.py /root/

ENV PYTHONPATH /root

CMD ["python3", "./run_app.py"]