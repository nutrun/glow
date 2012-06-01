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
        subprocess.check_call([glow_executable(), '-tube', 'listener_runs_job', '-out', tmpfilename, '/bin/echo', 'listener_runs_job'])
        self.wait_for_condition(lambda: os.path.exists(tmpfilename) and os.stat(tmpfilename).st_size > 0)
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('listener_runs_job\n', outfile.read())
        listener.send_signal(SIGINT)

    def test_submit_many_jobs(self):
        tmpfilename1 = temporary_file_name()
        tmpfilename2 = temporary_file_name()
        listener = start_listener()
        glow = subprocess.Popen([glow_executable()], stdin=subprocess.PIPE)
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
        tmpfilename = temporary_file_name()
        listener = start_listener()
        subprocess.check_call([glow_executable(), '-tube', 'listener_finishes_job_on_interrupt', '-out', tmpfilename, sibling_path('sleepthenecho'), '5',  'listener_finishes_job_on_interrupt'])
        listener.send_signal(SIGINT)
        self.wait_for_condition(lambda: os.path.exists(tmpfilename) and os.stat(tmpfilename).st_size > 0, seconds=10)
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('listener_finishes_job_on_interrupt\n', outfile.read())

    def test_listener_kills_job_on_kill(self):
        tmpfilename = temporary_file_name()
        listener = start_listener()
        subprocess.check_call([glow_executable(), '-tube', 'listener_kills_job_on_kill', '-out', tmpfilename, sibling_path('sleepthenecho'), '5',  'listener_kills_job_on_kill'])
        listener.terminate()
        listener.wait()
        with open(tmpfilename, 'r') as outfile:
            self.assertNotEqual('listener_kills_job_on_kill\n', outfile.read())
    
    def wait_for_condition(self, cond_f, seconds=3):
        end_time = time.time() + seconds
        while time.time() < end_time:
            if cond_f():
                return
            time.sleep(0.5)
        self.fail('timed out')


debug = True

def start_listener():
    return subprocess.Popen([glow_executable(), '-listen'], stderr=None if debug else open('/dev/null', 'r'))

def temporary_file_name():
    if debug:
        _, tmpfilename = tempfile.mkstemp()
        print 'temporary file:', tmpfilename
        return tmpfilename
    else:
        return tempfile.NamedTemporaryFile().name

def sibling_path(filename):
    return os.path.join(os.path.dirname(__file__), filename)
    

def glow_executable():
    return os.path.join(os.path.dirname(__file__), 'glow')
