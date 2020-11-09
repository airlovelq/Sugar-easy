FROM ubuntu:18.04
RUN apt-get update

RUN apt-get install -y python3
RUN apt-get install -y python3-pip

WORKDIR /root

RUN mkdir /root/.pip
COPY test_example/pip.conf /root/.pip

COPY test_example/requirements.txt /root/requirements.txt
RUN pip3 install -r /root/requirements.txt

COPY /test_example/test.py /root/test.py

ENV PYTHONPATH /root

CMD ["python3", "/root/test.py"]