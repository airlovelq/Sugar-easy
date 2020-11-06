FROM ubuntu:18.04

RUN apt-get update

RUN apt-get install -y python3
RUN apt-get install -y python3-pip

WORKDIR /root

RUN mkdir /root/.pip
COPY scripts/pip.conf /root/.pip

COPY src/* /root/
RUN pip3 install -r /root/requirements.txt

RUN mkdir /root/dataset
RUN mkdir /root/model
RUN mkdir /root/checkpoint
RUN mkdir /root/train
RUN mkdir /root/params
RUN mkdir /root/predict
RUN mkdir /root/predict/logs

ENV PYTHONPATH /root

CMD ["python3", "./run_app.py"]