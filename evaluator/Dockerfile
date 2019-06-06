FROM python:3.7.3

WORKDIR /evaluator

COPY evaluator .
RUN chmod 777 run.sh

RUN pip install -r requirements.txt

ENTRYPOINT ["bash", "/evaluator/run.sh"]
CMD []