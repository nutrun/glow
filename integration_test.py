import cjson
import os
import subprocess
import tempfile
import time
import unittest

SIGINT = 2

class TestGlowIntegration(unittest.TestCase):

    def test_listener_runs_job(self):
        tmpfilename = temporary_file_name()
        listener = start_listener()
        subprocess.check_call(['./glow', '-tube', 'listener_runs_job', '-out', tmpfilename, '/bin/echo', 'listener_runs_job'])
        self.wait_for_condition(lambda: os.path.exists(tmpfilename) and os.stat(tmpfilename).st_size > 0)
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('listener_runs_job\n', outfile.read())
        listener.send_signal(SIGINT)

    def test_submit_many_jobs(self):
        tmpfilename1 = temporary_file_name()
        tmpfilename2 = temporary_file_name()
        listener = start_listener()
        glow = subprocess.Popen(['./glow'], stdin=subprocess.PIPE)
        print >>glow.stdin, cjson.encode([{'cmd': 'echo submit_many_jobs', 'tube': 'submit_many_jobs', 'out': tmpfilename1 },
                                          {'cmd': 'echo submit_many_jobs', 'tube': 'submit_many_jobs', 'out': tmpfilename2 }])
        glow.stdin.close()
        self.wait_for_condition(lambda: os.path.exists(tmpfilename1) and os.stat(tmpfilename1).st_size > 0)
        with open(tmpfilename1, 'r') as outfile:
            self.assertEqual('submit_many_jobs\n', outfile.read())
        self.wait_for_condition(lambda: os.path.exists(tmpfilename2) and os.stat(tmpfilename2).st_size > 0)
        with open(tmpfilename2, 'r') as outfile:
            self.assertEqual('submit_many_jobs\n', outfile.read())
        listener.send_signal(SIGINT)

    def test_listener_finishes_job_on_interrupt(self):
        pass

    def test_listener_kills_job_on_kill(self):
        pass
    
    def wait_for_condition(self, cond_f):
        end_time = time.time() + 3 # seconds
        while time.time() < end_time:
            if cond_f():
                return
            time.sleep(0.5)
        self.fail('timed out')


debug = False

def start_listener():
    return subprocess.Popen(['./glow', '-listen'], stderr=None if debug else open('/dev/null', 'r'))

def temporary_file_name():
    if debug:
        _, tmpfilename = tempfile.mkstemp()
        print 'temporary file:', tmpfilename
        return tmpfilename
    else:
        return tempfile.NamedTemporaryFile().name
