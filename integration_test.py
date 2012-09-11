import cjson
import os
import subprocess
import tempfile
import unittest

from select import select

HERE = os.path.dirname(os.path.abspath(__file__))

class TestGlowIntegration(unittest.TestCase):

    def setUp(self):
        for tube in tubes():
            drain(tube)
        self.listener = Listener()
    
    def tearDown(self):
        self.listener.kill()

    def test_listener_runs_job(self):
        tmpfilename = temporary_file_name()
        self.listener.start()
        subprocess.check_call([glow_executable(), '-tube', 'listener_runs_job', '-out', tmpfilename, '/bin/echo', 'listener_runs_job'])
        self.listener.wait_for_job_completion({'tube': 'listener_runs_job', 'out': tmpfilename})
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('listener_runs_job\n', outfile.read())
        self.listener.interrupt()

    def test_local_runs_job(self):
        tmpfilename = temporary_file_name()
        subprocess.check_call([glow_executable(), '-local', '-out', tmpfilename, '/bin/echo', 'local_runs_job'])
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('local_runs_job\n', outfile.read())

    def test_submit_many_jobs(self):
        tmpfilename1 = temporary_file_name()
        tmpfilename2 = temporary_file_name()
        self.listener.start()

        glow = subprocess.Popen([glow_executable()], stdin=subprocess.PIPE)
        print >>glow.stdin, cjson.encode([
                    {'cmd': 'echo', 'args': ['submit_many_jobs'], 'tube': 'submit_many_jobs', 'out': tmpfilename1 },
                    {'cmd': 'echo', 'args': ['submit_many_jobs'], 'tube': 'submit_many_jobs', 'out': tmpfilename2 }
                ])
        glow.stdin.close()
        self.listener.wait_for_job_completion({'tube': 'submit_many_jobs', 'out': tmpfilename1})
        self.listener.wait_for_job_completion({'tube': 'submit_many_jobs', 'out': tmpfilename2})

        with open(tmpfilename1, 'r') as outfile:
            self.assertEqual('submit_many_jobs\n', outfile.read())
        with open(tmpfilename2, 'r') as outfile:
            self.assertEqual('submit_many_jobs\n', outfile.read())
        self.listener.interrupt()

    def test_listener_finishes_job_on_interrupt(self):
        tmpfilename = temporary_file_name()
        self.listener.start()

        subprocess.check_call([glow_executable(), '-tube', 'listener_finishes_job_on_interrupt', '-out', tmpfilename, '%s/sleepthenecho' % HERE, '3',  'listener_finishes_job_on_interrupt'])
        self.listener.wait_for_job_start({'tube': 'listener_finishes_job_on_interrupt', 'out': tmpfilename})

        self.listener.interrupt()
        self.listener.wait_for_job_completion({'tube': 'listener_finishes_job_on_interrupt', 'out': tmpfilename}, seconds=10)

        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('listener_finishes_job_on_interrupt\n', outfile.read())

    def test_listener_kills_job_on_kill(self):
        tmpfilename = temporary_file_name()
        self.listener.start()
        
        subprocess.check_call([glow_executable(), '-tube', 'listener_kills_job_on_kill', '-out', tmpfilename, '%s/sleepthenecho' % HERE, '5',  'listener_kills_job_on_kill'])
        self.listener.wait_for_job_start({'tube': 'listener_kills_job_on_kill', 'out': tmpfilename})
        
        self.listener.kill()
        self.listener.wait_for_shutdown()
        
        try:
            with open(tmpfilename, 'r') as outfile:
                self.assertNotEqual('listener_kills_job_on_kill\n', outfile.read())
        except IOError as e:
            # ignore if file was never created
            if e.errno != 2: raise

    def test_unexecable_job_fails_with_error(self):
        self.assertFalse('GLOW_ERRORS' in tubes())
        self.listener.start()
        subprocess.check_call([glow_executable(), '-tube', 'unexecable_job_fails_with_error', '/nonexistent/executable'])
        self.listener.wait_for_job_failure({'tube': 'unexecable_job_fails_with_error', 'cmd': '/nonexistent/executable'})
        self.assertEqual(1, tubes()['GLOW_ERRORS']['jobs-ready'])
    
    def test_nonzero_exitstatus_fails_with_error(self):
        self.assertFalse('GLOW_ERRORS' in tubes())
        self.listener.start()
        subprocess.check_call([glow_executable(), '-tube', 'nonzero_exitstatus_fails_with_error', 'cat', '/nonexistent/file'])
        self.listener.wait_for_job_failure({'tube': 'nonzero_exitstatus_fails_with_error', 'cmd': 'cat',
                                            'args': ['/nonexistent/file']})
        self.assertEqual(1, tubes()['GLOW_ERRORS']['jobs-ready'])

    def test_local_job_failure(self):
        self.assertFalse('GLOW_ERRORS' in tubes())
        returncode = subprocess.call([glow_executable(), '-local', 'cat', '/nonexistent/file'], stderr=open('/dev/null', 'w'))
        self.assertNotEqual(0, returncode)
        self.assertEqual(1, tubes()['GLOW_ERRORS']['jobs-ready'])


    def test_unexecable_local_job_failure(self):
        self.assertFalse('GLOW_ERRORS' in tubes())
        returncode = subprocess.call([glow_executable(), '-local', '/nonexistent/executable'], stderr=open('/dev/null', 'w'))
        self.assertNotEqual(0, returncode)
        self.assertEqual(1, tubes()['GLOW_ERRORS']['jobs-ready'])

    def test_create_output_file_if_not_exists(self):
        tmpfilename = temporary_file_name()
        self.assertFalse(os.path.exists(tmpfilename))
        self.listener.start()
        subprocess.check_call([glow_executable(), '-tube', 'job', '-out', tmpfilename, '/bin/echo', 'job'])
        self.listener.wait_for_job_completion({'tube': 'job', 'out': tmpfilename})
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('job\n', outfile.read())
        self.listener.interrupt()

    def test_append_to_output_file_if_exists(self):
        tmpfilename = temporary_file_name()
        self.listener.start()
        subprocess.check_call([glow_executable(), '-tube', 'job1', '-out', tmpfilename, '/bin/echo', 'job1'])
        subprocess.check_call([glow_executable(), '-tube', 'job2', '-out', tmpfilename, '/bin/echo', 'job2'])
        self.listener.wait_for_job_completion({'tube': 'job2', 'out': tmpfilename})
        with open(tmpfilename, 'r') as outfile:
            self.assertEqual('job1\njob2\n', outfile.read())
        self.listener.interrupt()
    

debug = False

class Listener:
    def __init__(self):
        self.process = None

    def start(self):
        self.process = subprocess.Popen([glow_executable(), '-listen', '-v'], stderr=subprocess.PIPE)

    def interrupt(self):
        # Send SIGINT
        self.process.send_signal(2) 

    def kill(self):
        if self.process:
            try:
                self.process.terminate()
            except OSError as e:
                # ignore if 'No such process' (already killed)
                if e.errno != 3:
                    raise

    def wait_for_shutdown(self):
        self.process.wait()

    def wait_for_job_start(self, job_desc, seconds=3):
        self._wait_for_job_update(job_desc, 'RUNNING:', seconds)

    def wait_for_job_completion(self, job_desc, seconds=3):
        self._wait_for_job_update(job_desc, 'COMPLETE:', seconds)

    def wait_for_job_failure(self, job_desc, seconds=3):
        self._wait_for_job_update(job_desc, 'FAILED:', seconds)

    def _wait_for_job_update(self, job_desc, status, seconds, max_num_non_matching_events=10):
        num_events = 0
        while num_events < max_num_non_matching_events:
            fds, _, _ = select([self.process.stderr], [], [], seconds)
            if fds != [self.process.stderr]:
                raise Exception('timed out waiting for {0} {1}'.format(status, job_desc))
            line = self.process.stderr.readline()
            if debug: print line
            if line.startswith(status):
                job = cjson.decode(line[len(status):])
                if all([job[k] == job_desc[k] for k in job_desc]):
                    return job
            num_events += 1


def temporary_file_name():
    if debug:
        _, tmpfilename = tempfile.mkstemp()
        print 'temporary file:', tmpfilename
        return tmpfilename
    else:
        return tempfile.NamedTemporaryFile().name

def glow_executable():
    return '%s/glow' % HERE

def tubes():
    return cjson.decode(subprocess.check_output([glow_executable(), '-stats']))

def drain(tube):
    subprocess.check_call([glow_executable(), '-drain', tube])


    
if __name__ == '__main__':
    # this works, but nose is better
    unittest.main()
