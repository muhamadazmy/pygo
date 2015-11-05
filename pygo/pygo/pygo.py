import os
import struct
import json
import imp
import traceback


CHANNEL_IN = 3
CHANNEL_OUT = 4

HEADER_FMT = '>I'
HEADER_SIZE = struct.calcsize(HEADER_FMT)


def readlen(f, n):
    buffer = ''
    while len(buffer) < n:
        data = f.read(n - len(buffer))
        if data == '':
            raise Exception('EOF')

        buffer += data

    return buffer


class Runner(object):
    def __init__(self, module):
        self.module = module

        # open channel files
        self.chan_in = os.fdopen(CHANNEL_IN, 'r')
        self.chan_out = os.fdopen(CHANNEL_OUT, 'w')

        self.mod = None

    def get_next_call(self):
        header = readlen(self.chan_in, HEADER_SIZE)
        length = struct.unpack(HEADER_FMT, header)[0]

        data = readlen(self.chan_in, length)

        return json.loads(data)

    def send_result(self, result):
        data = json.dumps(result)
        self.chan_out.write(struct.pack(HEADER_FMT, len(data)))
        self.chan_out.write(data)
        self.chan_out.flush()

    def get_module(self):
        if self.mod is None:
            self.mod = imp.load_module(self.module, *imp.find_module(self.module))

        return self.mod

    def do_call(self, call):
        module = self.get_module()
        func_name = call['function']
        args = call['kwargs']

        result = {}
        try:
            func = getattr(module, func_name)
            call_result = func(**args)
            result = {
                'return': call_result
            }
        except:
            result = {
                'state': 'ERROR',
                'return': traceback.format_exc()
            }

        return result

    def run(self):
        while True:
            result = {'state': 'ERROR'}
            try:
                call = self.get_next_call()
                result = self.do_call(call)
            except:
                result['return'] = traceback.format_exc()
            finally:
                self.send_result(result)
