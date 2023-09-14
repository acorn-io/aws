FROM python:3-alpine

WORKDIR /src

ENV FLASK_APP=app.py
ENV FLASK_RUN_HOST=0.0.0.0

ADD . /src
RUN pip install -r requirements.txt

COPY . .
EXPOSE 5000
CMD ["flask", "--debug", "run"]