FROM python:3.6.8-stretch

WORKDIR /aggregator
COPY . .

RUN pip install -r requirements.txt

RUN chmod 777 run.sh \
    && ln -s run.sh /

ENTRYPOINT ["bash", "run.sh"]
CMD []