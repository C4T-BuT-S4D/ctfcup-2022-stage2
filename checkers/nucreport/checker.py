#!/usr/bin/env python3
import os
import string
import sys

import requests
from checklib import *

import reports_lib


class Checker(BaseChecker):
    vulns: int = 1
    timeout: int = 15
    uses_attack_data: bool = True

    def __init__(self, *args, **kwargs):
        super(Checker, self).__init__(*args, **kwargs)
        self.cm = reports_lib.CheckMachine(self)

    def action(self, action, *args, **kwargs):
        try:
            super(Checker, self).action(action, *args, **kwargs)
        except requests.exceptions.ConnectionError:
            self.cquit(Status.DOWN, 'Connection error', 'Got requests connection error')

    def random_file(self):
        rnd_filename = rnd_string(5) + '.txt'
        rnd_content = rnd_string(50)
        return rnd_filename, rnd_content.encode()

    def docx_file(self):
        with open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'fake_data/test.docx'), 'rb') as f:
            return 'report_' + rnd_string(3) + '.docx', f.read()

    def random_username(self):
        u = rnd_username(salt_length=0) + rnd_string(3, alphabet=string.digits + string.ascii_lowercase)
        while len(u) < 6:
            u = rnd_username(salt_length=0) + rnd_string(3, alphabet=string.digits + string.ascii_lowercase)
        return u

    def check(self):
        sess = get_initialized_session()
        u = self.random_username()
        p = rnd_password(10)

        self.cm.register(sess, u, p)
        self.cm.login(sess, u, p, status=status.Status.MUMBLE)

        check_files = (self.random_file(), self.random_file())
        for filename, content in check_files:
            self.cm.put_file(u, p, filename, content)

            self.assert_eq(content, self.cm.get_file(u, p, filename), 'Failed to get file using SFTP',
                           status.Status.MUMBLE)

        self.cm.reindex_files(sess)

        files_indexed = self.cm.list_user_files_web(sess)

        new_sess = get_initialized_session()
        su = self.random_username()
        sp = rnd_password(10)
        self.cm.register(new_sess, su, sp)
        self.cm.login(new_sess, su, sp, status=status.Status.MUMBLE)

        for f in files_indexed:
            path = f['path']
            token = f['token']
            f_content = ''
            for fname, content in check_files:
                if fname in path:
                    f_content = content
                    break

            self.assert_eq(self.cm.get_file_web(sess, path, token=''), f_content, "Failed to retrieve user-file")
            self.assert_eq(self.cm.get_file_web(new_sess, path, token=token), f_content,
                           'Failed to retrieve file by token')

        self.cquit(Status.OK)

    def put(self, flag_id: str, flag: str, vuln: str):
        sess = get_initialized_session()
        u = self.random_username()
        p = rnd_password(10)

        self.cm.register(sess, u, p)
        self.cm.login(sess, u, p, status=status.Status.MUMBLE)
        self.cm.put_file(u, p, 'flag.txt', flag.encode())
        self.cm.reindex_files(sess)

        files_indexed = self.cm.list_user_files_web(sess)

        flags_indexed = [x for x in files_indexed if 'flag.txt' in x['path']]
        self.assert_gte(len(flags_indexed), 1, "Failed to find uploaded file link at the home page")

        flag_info = flags_indexed[0]

        self.cquit(Status.OK, flag_info['path'], f"{u}:{p}:{flag_info['path']}:{flag_info['token']}")

    def get(self, flag_id: str, flag: str, vuln: str):
        s = get_initialized_session()
        username, password, path, token = flag_id.split(':')

        self.cm.login(s, username, password, status=status.Status.CORRUPT)

        self.assert_eq(self.cm.get_file(username, password, 'flag.txt').decode(), flag, 'Failed to get file using SFTP',
                       status=status.Status.CORRUPT)

        self.assert_eq(self.cm.get_file_web(s, path, token, status=status.Status.CORRUPT).decode(), flag,
                       'Failed to retrieve file content using app', status=status.Status.CORRUPT)

        self.cquit(Status.OK)


if __name__ == '__main__':
    c = Checker(sys.argv[2])

    try:
        c.action(sys.argv[1], *sys.argv[3:])
    except c.get_check_finished_exception():
        cquit(Status(c.status), c.public, c.private)
