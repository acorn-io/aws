import mimetypes
import redis
import os

from types import NoneType
from flask import Flask, render_template, request

from natsort import natsorted, ns

redis_pass=os.getenv('REDIS_PASSWORD')
app = Flask(__name__)
cache = redis.Redis(host=os.getenv('REDIS_HOST'), port=6379, db=0, password=redis_pass, decode_responses=True, ssl=True)

@app.route("/", methods=["GET", "POST"])
def index():
    text=""
    if request.method == 'POST':
        text=request.form['echotext']

        current_message_count = cache.incr('msgctr')
        key = "message" + str(current_message_count)

        cache.set(key, text)
    
    keys = get_all_redis_keys('message')
    messages = get_redis_kv_map(keys)
    
    return render_template("index.html.tmpl", text=text, messages=messages)


def get_redis_kv_map(keys):
    kv_map = {}
    for key in keys:
        kv_map[key] = cache.get(key)
    return kv_map

def get_all_redis_keys(prefix):
    return cache.keys("%s*" % prefix)

@app.template_filter()
def natural_sort(key):
    return natsorted(key, reverse=True, alg=ns.IGNORECASE)