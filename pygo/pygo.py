import os
import struct
import json
import imp
import traceback
import sys


CHANNEL_IN = 3
CHANNEL_OUT = 4

HEADER_FMT = '>I'
HEADER_SIZE = struct.calcsize(HEADER_FMT)


def readlen(f, n):
    data = ''
    while len(data) < n:
        data += f.read(n - len(data))

    return data


def get_next_call(chan_in):
    header = readlen(chan_in, HEADER_SIZE)
    length = struct.unpack(HEADER_FMT, header)[0]

    data = readlen(chan_in, length)

    return json.loads(data)


def do_call(module, call):
    func_name = call['function']
    args = call['kwargs']

    result = {}
    try:
        func = getattr(module, func_name)
        call_result = func(**args)
        result = {
            'return': call_result
        }
    except Exception, e:
        result = {
            'state': 'ERROR',
            'return': str(e)
        }

    return result


def send_result(chan_out, result):
    data = json.dumps(result)
    chan_out.write(struct.pack(HEADER_FMT, len(data)))
    chan_out.write(data)
    chan_out.flush()


def run(module):
    # open channel files
    chan_in = os.fdopen(CHANNEL_IN, 'r')
    chan_out = os.fdopen(CHANNEL_OUT, 'w')

    mod = imp.load_module(module, *imp.find_module(module))
    try:
        while True:
            call = get_next_call(chan_in)
            result = do_call(mod, call)
            send_result(chan_out, result)
    except:
        traceback.print_exc(file=sys.stderr)
