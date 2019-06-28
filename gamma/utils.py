import os
from os import path as pth
from itertools import tee
from atexit import register
from tempfile import _get_default_tempdir, _get_candidate_names
from shutil import copyfileobj
from argparse import ArgumentParser
from glob import iglob
from logging import getLogger


log = getLogger("gamma.sentinel1")

import gamma as gm


__all__ = [
    "all_same",
    "make_object",
    "make_join",
    "Params",
    "get_tmp",
    "get_par",
    "cat",
    "Files",
    "Multi",
    "Base",
    "Date",
    "CParse",
    "annot",
    "pos",
    "opt",
    "flag",
    "rm",
    "ln",
    "mv",
    "mkdir",
]


tmpdir = _get_default_tempdir()


def all_same(iterable):
    return len(set(tee(iterable,1))) == 1


def make_object(name, inherit=(object,), **kwargs):
    return type(name, inherit, **kwargs)



class Params(object):
    def __init__(self, dictionary):
        self.params = dictionary
    
    @classmethod
    def from_file(cls, path, sep=":"):
        with open(path, "r") as f:
            return cls({
                line.split(sep)[0].strip() : " ".join(line.split(sep)[1:]).strip()
                for line in f if line
            })
    
    def __str__(self):
        return pformat(self.params)
    
    def __getitem__(self, key):
        return self.params[key]

    def getfloat(self, key, idx=0):
        return float(self[key].split()[idx])

    def getint(self, key, idx=0):
        return int(self[key].split()[idx])


tmp = []


def get_tmp(path=tmpdir):
    global tmp
    path = pth.join(path, next(_get_candidate_names()))
    
    tmp.append(path)
    
    return path


def cleanup_temps():
    global tmp
    
    for path in tmp:
        log.debug('Removed file: "%s"' % path)
        rm(path)


register(cleanup_temps)


def cat(out, *args):
    assert len(args) >= 1, "Minimum one input file is required"
    
    with open(out, 'wb') as f_out, open(args[0], 'rb') as f_in:
        copyfileobj(f_in, f_out)

    for arg in args[1:]:
        with open(out, 'ab') as f_out, open(arg, 'rb') as f_in:
            copyfileobj(f_in, f_out)


class Files(object):
    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, value)
    
    def stat(self, attrib, rng=None, roff=0, loff=0, nr=None, nl=None):
        obj = getattr(self, attrib)
        
        if rng is None:
            rng = obj.rng()
        
        if isinstance(obj, string_t):
            gp.image_stat(obj, rng, roff, loff, nr, nl, "tmp")
        else:
            gp.image_stat(obj.dat, rng, roff, loff, nr, nl, "tmp")
        
        pars = Params.from_file("tmp")
        rm("tmp")
        
        return pars
    

    def exist(self, *attribs):
        return all(pth.isfile(getattr(self, attrib)) for attrib in attribs)
    
    
    def mv(self, attrib, dst):
        Files._mv(getattr(self, attrib), dst)
    
    
    def move(self, attribs, dst):
        for attrib in attribs:
            attr = getattr(self, attrib)
            Files._mv(attr, dst)
            
            newpath = pth.join(pth.abspath(dst), pth.basename(attr))
            
            setattr(self, attrib, newpath)
    
    
    def rm(self, *attribs):
        for attrib in attribs:
            rm(getattr(self, attrib))
    
    def ln(self, attrib, other):
        ln(getattr(self, attrib), other)
    
    def cp(self, attrib, other):
        sh.copy(getattr(self, attrib), other)
    
    # def get(self, attrib, key):
    #     return Files.get_par(key, getattr(self, attrib))
    # 
    # def getfloat(self, attrib, key, idx=0):
    #     return Files._getfloat(key, getattr(self, attrib), idx)
    # 
    # def getint(self, attrib, key, idx=0):
    #     return Files._getint(key, getattr(self, attrib), idx)
    
    def set(self, attrib, key, **kwargs):
        return Files.set_par(key, getattr(self, attrib), **kwargs)
    
    
    def empty(self, attrib):
        return Files.is_empty(getattr(self, attrib))

    
    @staticmethod
    def _mv(src, dst):
        if pth.isfile(dst):
            dst_ = dst
        elif pth.isdir(dst):
            dst_ = pth.join(dst, pth.basename(src))
            
        rm(dst_)
        sh.move(src, dst_)
        
        log.debug('File "%s" moved to "%s".' % (src, dst_))


    @staticmethod
    def is_empty(path):
        return pth.getsize(path) == 0
    
    

        
    @staticmethod
    def _getfloat(key, data, idx=0):
        return float(Files.get_par(key, data).split()[idx])

    @staticmethod
    def _getint(key, data, idx=0):
        return int(Files.get_par(key, data).split()[idx])


    @staticmethod
    def set_par(key, infile, new=""):
        if Files.is_empty(infile):
            with open(infile, "w") as f:
                f.write("%s: %s\n" % (key, new))
            
            return
        
        
        with open(infile, "r+") as f:
            lines = (line for line in f)
        
            lines = (
                        "%s: %s" % (key, new)
                        if key in line
                        else line
                        for line in lines
                    )
            
            f.seek(0)
            f.truncate()
            
            f.write("%s\n" % "\n".join(lines))


def get_par(key, data, sep=":"):
    if pth.isfile(data):
        with open(data, "r") as f:
            lines = f.readlines()
    elif isinstance(data, bytes):
        lines = data.decode().split("\n")
    elif isinstance(data, string_t):
        lines = data.split("\n")
    else:
        lines = data
    
    for line in lines:
        if key in line:
            return " ".join(line.split(sep)[1:]).strip() 

    
def Multi(**kwargs):
    return type("Multi", (object,), kwargs)


class Base(Files):
    def __init__(self, base, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, "%s%s" % (base, value))
        
        self.base = base
    
    def rm(self):
        for elem in dir(self):
            Files.rm(self, elem)


class Date(object):
    __slots__ = ("start", "stop", "center")
    
    def __init__(self, start_date, stop_date, center=None):
        self.start = start_date
        self.stop = stop_date
        
        if center is None:
            center = (start_date - stop_date) / 2.0
            center = stop_date + center
    
        self.center = center
    
    
    def date2str(self, fmt="%Y%m%d"):
        return self.center.strftime(fmt)
    
    def __eq__(self, other):
        return (self.start == other.start and self.stop == other.stop and
                self.mean == other.mean)
    
    def __str__(self):
        return self.date2str()

    def __repr__(self):
        return "<Date start: %s stop: %s mean: %s>"\
                % (self.start, self.stop, self.mean)

def mkdir(path):
    try:
        os.makedirs(path)
        log.debug("Directory \"{}\" created.".format(path))
        return path
    except OSError as e:
        if e.errno != EEXIST:
            raise e
        else:
            log.debug("Directory \"{}\" already exists.".format(path))
            return path


def ln(target, link_name):
    try:
        os.symlink(target, link_name)
    except OSError as e:
        if e.errno == EEXIST:
            os.remove(link_name)
            os.symlink(target, link_name)
            log.debug("Symlink from \"%s\" to \"%s\" created"
                         % (target, link_name))
        else:
            raise e


def rm(*args):
    for arg in args:
        for path in iglob(arg):
            if not pth.isfile(path) and not pth.isdir(path):
                return
            elif pth.isdir(path):
                sh.rmtree(path)
                log.debug("Directory \"%s\" deleted," % path)
            elif pth.isfile(path):
                try:
                    os.remove(path)
                    log.debug("File \"%s\" deleted." % path)
                except OSError as e:
                    if e.errno != ENOENT:
                        raise e
            else:
                raise Exception("%s is not a file nor is a directory!" % path)


def mv(*args, **kwargs):
    dst = kwargs.pop("dst", None)
    for arg in args:
        for src in iglob(arg): 
            rm(pth.join(dst, src))
            sh.move(src, dst)
            log.debug("File \"%s\" moved to \"%s\"." % (src, dst))    

            
def make_join(path):
    def f(*args, **kwargs):
        return pth.join(path, *args, **kwargs)
    
    return f


def pos(action="store", help=None, type=str, choices=None,
        nargs=None, metavar=None, dest=None, const=None):
    return {
        "action": action,
        "nargs": nargs,
        "type": type,
        "choices": choices,
        "help": help,
        "metavar": metavar,
        "kind": "pos"
        # "dest": dest,
    }


def opt(short=None, action="store", help=None, type=str, choices=None,
        nargs=None, metavar=None, dest=None, default=None, const=None):
    
    ret = {
        "action": action,
        "nargs": nargs,
        "default": default,
        "type": type,
        "choices": choices,
        "required": False,
        "help": help,
        "metavar": metavar,
        "dest": dest,
        "nargs": nargs,
        "kind": "opt"
    }
    
    
    # if short is not None:
    #     ret["flags"] = "-" + short
    
    return ret


def flag(short=None, action="store_true", help=None, dest=None):

    ret = {
        "action": action,
        "help": help,
        "dest": dest,
        "kind": "flag"
    }
    
    
    # if short is not None:
    #     ret["flags"] = "-" + short
    
    return ret


def annot(**kwargs):
    parent = kwargs.pop("parent", None)
    
    def annotate(f):
        f.__parent__ = parent
        f.__annotations__ = kwargs
        
        return f
        
    return annotate


class CParse(object):
    def __init__(self, **kwargs):
        self.argp, self.args = ArgumentParser(**kwargs), None
        
        for key, value in self.__init__.__annotations__.items():
            if value.pop("kind") == "pos":
                self.argp.add_argument(key, **value)
            else:
                self.argp.add_argument("--" + key, **value)
            
        
        self.subparser = self.argp.add_subparsers(**kwargs)
        
        try:
            cmds = self.commands
        except AttributeError:
            return
        
        for cmd in cmds:
            fun = getattr(self, cmd)
            subp = self.subparser.add_parser(fun.__name__)
            
            try:
                parent = fun.__parent__
            except AttributeError:
                parent = None
            
            if parent is not None:
                fun.__annotations__.update(getattr(self, parent).__annotations__)
            
            
            for key, value in fun.__annotations__.items():
                if value.pop("kind") == "pos":
                    subp.add_argument(key, **value)
                else:
                    subp.add_argument("--" + key, **value)
            
            subp.set_defaults(fun=fun)
    
    
    def __getitem__(self, item):
        return getattr(self.args, item)
    
    
    def parse(self):
        self.args = self.argp.parse_args()
        return self
