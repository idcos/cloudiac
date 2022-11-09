# coding=utf-8

from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = '''
  name: saltapi
  short_description: Ansible connection plugin for salt-api
  author: Jinxing <jinxing@idcos.com>
  options:
    host:
      description: Salt Minion Id to connect to.
      vars:
        - name: inventory_hostname
        - name: ansible_host

    endpoint: 
      description: Salt-api endpoint.
      env:
        - name: ANSIBLE_SALTAPI_ENDPOINT
      vars:
        - name: ansible_saltapi_endpoint
      required: True

    username: 
      description: Salt-api username.
      env:
        - name: ANSIBLE_SALTAPI_USERNAME
      vars:
        - name: ansible_saltapi_username
      required: True

    password: 
      description: Salt-api password.
      env:
        - name: ANSIBLE_SALTAPI_PASSWORD
      vars:
        - name: ansible_saltapi_password
      required: True

    eauth: 
      description: Salt-api eauth type.
      default: 'pam'
      env:
        - name: ANSIBLE_SALTAPI_EAUTH
      vars:
        - name: ansible_saltapi_eauth
'''

import os
import base64
import json
import sys
import ssl

if sys.version_info.major == "3":
    from urllib.request import Request, urlopen, HTTPError
else:
    from urllib2 import Request, urlopen, HTTPError

from ansible.errors import (
    AnsibleAuthenticationFailure,
    AnsibleConnectionFailure,
    AnsibleError,
)
from ansible.plugins.connection import ConnectionBase


def request(url, data, headers):
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    req = Request(url, data, headers)

    try:
      resp = urlopen(req, context=ctx)
    except HTTPError as e:
        if e.code == 401:
            raise AnsibleAuthenticationFailure(e.reason)
        else:
            raise AnsibleConnectionFailure(e.reason)
    return resp.read()


class Connection(ConnectionBase):
    has_pipelining = False
    transport = 'saltapi'

    def __init__(self, play_context, *args, **kwargs):
        super(Connection, self).__init__(play_context, *args, **kwargs)
        self.host = None
        self.endpoint = None
        self.username = None
        self.password = None
        self.eauth = None
        self.session = None
        self.token = None

    def _post(self, path, jsondata=None):
        url = "{0}{1}".format(self.endpoint.rstrip("/"), path) 
        self._display.vvv("POST to salt-api: %s" % url)
        headers={
            "Accept": "application/json",
            "Content-Type": "application/json"
        }

        # 经测试添加 X-Auth-Token 无效，依然报 401，
        # 所以改成了在请求 body 中直接提供 username+password
        # if self.token:
        #     headers["X-Auth-Token"] = self.token

        if jsondata is None:
            jsondata = {}

        self._display.vvvvv("POST Headers: %s" % headers)
        body = json.dumps(jsondata, ensure_ascii=False)
        self._display.vvvvv("POST Body: %s" % body)
        resp = request(url, body.encode("utf8"), headers=headers)
        self._display.vvvvv("Salt-api response: %s" % resp)
        result = json.loads(resp)
        self._display.vvvv("Salt-api return: %s" % result["return"])
        return result["return"]

    def _connect(self):
        if not self._connected:
            self.host = self.get_option('host') or self._play_context.remote_addr
            self.endpoint = self.get_option('endpoint')
            self.username = self.get_option('username')
            self.password = self.get_option('password')
            self.eauth = self.get_option('eauth')

            r = self._post("/login", jsondata={
                "username": self.username,
                "password": self.password,
                "eauth": self.eauth
            })
            self.token = r[0]["token"]
            self._connected = True
        return self

    def _run_local(self, fun, arg=None, kwarg=None):
        data = {
            "client": "local",
            "tgt": self.host,
            "fun": fun,
            "username": self.username,
            "password": self.password,
            "eauth": self.eauth
        }
        if arg:
            data["arg"] = arg
        if kwarg:
            data["kwarg"] = kwarg
        result = self._post("/run", jsondata=[data])[0]
        if self.host not in result:
            raise AnsibleError("Minion %s didn't answer, check if salt-minion is running and the name is correct" % self.host)
        return result[self.host]

    def exec_command(self, cmd, in_data=None, sudoable=True):
        """ run a command on the remote minion """
        super(Connection, self).exec_command(cmd, in_data=in_data, sudoable=sudoable)

        if in_data:
            raise AnsibleError("Internal Error: this module does not support optimized module pipelining")

        self._display.vvv("EXEC %s" % cmd, host=self.host)
        # need to add 'true;' to work around https://github.com/saltstack/salt/issues/28077
        r = self._run_local('cmd.exec_code_all', arg=['bash', 'true;' + cmd])
        return r['retcode'], r['stdout'], r['stderr']

    @staticmethod
    def _normalize_path(path, prefix):
        if not path.startswith(os.path.sep):
            path = os.path.join(os.path.sep, path)
        normpath = os.path.normpath(path)
        return os.path.join(prefix, normpath[1:])

    def put_file(self, in_path, out_path):
        """ transfer a file from local to remote """

        super(Connection, self).put_file(in_path, out_path)

        out_path = self._normalize_path(out_path, '/')
        self._display.vvv("PUT %s TO %s" % (in_path, out_path), host=self.host)
        with open(in_path, 'rb') as in_fh:
            content = in_fh.read()
        r = self._run_local('hashutil.base64_decodefile', kwarg={
            "instr": base64.b64encode(content).decode("utf8"), 
            "outfile": out_path
        })
        if r != True:
          raise AnsibleError("PUT %s TO %s: %s" % (in_path, out_path, r))

    def fetch_file(self, in_path, out_path):
        """ fetch a file from remote to local """

        super(Connection, self).fetch_file(in_path, out_path)

        in_path = self._normalize_path(in_path, '/')
        self._display.vvv("FETCH %s TO %s" % (in_path, out_path), host=self.host)
        r = self._run_local('hashutil.base64_encodefile', [in_path])
        with open(out_path, 'wb') as fp:
            fp.write(base64.b64decode(r))

    def close(self):
        """ terminate the connection; nothing to do here """
        pass
