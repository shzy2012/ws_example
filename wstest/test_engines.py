#!/usr/bin/python
#-*- encoding:utf-8 -*-

from __future__ import print_function
from __future__ import unicode_literals
from ws4py.client.threadedclient import WebSocketClient
from multiprocessing import Queue
import argparse
import io
import json
import logging
import sys
import threading
import time


FREQ = 4


def rate_limited(maxPerSecond):
    minInterval = 1.0 / float(maxPerSecond)

    def decorate(func):
        lastTimeCalled = [0.0]
        def rate_limited_function(*args,**kargs):
            elapsed = time.process_time() - lastTimeCalled[0]
            leftToWait = minInterval - elapsed
            if leftToWait > 0:
                time.sleep(leftToWait)
            ret = func(*args,**kargs)
            lastTimeCalled[0] = time.process_time()
            return ret
        return rate_limited_function

    return decorate


def format_nbest(asr_response_json):
    return '\n'.join([str(hyp) for hyp in asr_response_json['result']['hypotheses']])


class KzyAsrWebSocketClient(WebSocketClient):
    def __init__(self,
                 audiofile,
                 url,
                 protocols=None,
                 extensions=None,
                 heartbeat_freq=None,
                 byterate=16000):  # 16kHz x 16bit (2byte) = 32000byte/sec
        super(KzyAsrWebSocketClient, self).__init__(
            url, protocols, extensions, heartbeat_freq)
        self.final_hyps = []
        self.audiofile = audiofile
        self.byterate = byterate
        self.final_hyp_queue = Queue()

    @rate_limited(FREQ)
    def send_data(self, data):
        self.send(data, binary=True)

    def opened(self):
        logging.info('WebSocket opened().')

        def send_data_to_ws():
            with self.audiofile as audiostream:
                while True:
                    block = audiostream.read(int(self.byterate / FREQ))
                    if len(block) == 0:
                        break
                    self.send_data(block)

            logging.info('Audio sent, now sending EOS.')
            self.send('EOS')

        t = threading.Thread(target=send_data_to_ws)
        t.start()

    def received_message(self, msg):
        response = json.loads(str(msg))
        logging.info('response: %s', response)
        if response['status'] == 0:
            if 'result' in response:
                trans = response['result']['hypotheses'][0]['transcript']
                if response['result']['final']:
                    self.final_hyps.append(trans)
                    logging.info('[句尾] %s', trans.replace('\n', '\\n'))
                    logging.info('[nbest] %s', format_nbest(response))
                else:
                    print_trans = trans.replace('\n', '\\n')
                    logging.info('[句中] %s', print_trans)
        else:
            logging.error('Received server error with status [%d].',
                          response['status'])
            if 'message' in response:
                logging.error('Error message [%s].', response['message'])

    def get_full_hyp(self, timeout=30):
        return self.final_hyp_queue.get(timeout)

    def closed(self, code, reason=None):
        logging.info('WebSocket closed().')
        self.final_hyp_queue.put(' '.join(self.final_hyps))


def run(audio, byterate, engine_url, transcript):
    ws = KzyAsrWebSocketClient(audio, engine_url, byterate=byterate)
    ws.connect()
    result = ws.get_full_hyp()
    with io.open(transcript, 'w', encoding='utf-8') as t:
        t.write(result)
    logging.info('Final result [%s].', result)


def main(argv):
    logging.basicConfig(level=logging.INFO,
                        format='%(asctime)s %(name)s %(levelname)-8s %(message)s')
    parser = argparse.ArgumentParser(description='客知音python脚本模板')

    parser.add_argument('--audio',
                        help='待转录的音频文件',
                        type=argparse.FileType('rb'),
                        required=True)
    parser.add_argument('--byterate',
                        help='音频文件byterate，单位是KB/s',
                        type=int,
                        default=16000)
    parser.add_argument('--transcript',
                        help='输出的转录结果JSON文件',
                        type=str,
                        default='/tmp/transcript.json')
    parser.add_argument('--url',
                        help='''引擎URL''',
                        type=str,
                        default='ws://0.0.0.0:18080/asr/streaming')
    args = parser.parse_args()

    logging.info(args)

    run(args.audio, args.byterate, args.url, args.transcript)


if __name__ == '__main__':
    main(sys.argv)
