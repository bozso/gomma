from subprocess import Popen, STDOUT, PIPE

import gamma as gm

__all__ = ("gset", "histo", "cmd", "output", "title", "plot", "term",
           "histo", "get_format")


formats = {
    "FCOMPLEX" : "%float32%float32",
    "SCOMPLEX" : "%short%short",
    "FLOAT"    : "%float32",
    "INT"      : "%int",
    "SHORT"    : "%short",
    "BYTE"     : "%byte"
}


__plot_cmds = ""


def cmd(*args):
    global __plot_cmds
    __plot_cmds = "%s\n%s" % (__plot_cmds, "\n".join(args))


def plot(**kwargs):
    global __plot_cmds
    debug   = bool(kwargs.get("debug", False))
    persist = bool(kwargs.get("persist", False))
    

    if debug:
        print(__plot_cmds)
        return 1
    
    if persist:
        cmd = ["gnuplot", "-persist"]
    else:
        cmd = ["gnuplot"]
    
    proc = Popen(cmd, stderr=STDOUT, stdin=PIPE)
    proc.communicate(input=__plot_cmds.encode("ascii"))


def parse_set(key, value):
    if value is True:
        return "set {}".format(key)
    elif value is False:
        return "unset {}".format(key)
    else:
        return "set {} '{}'".format(key, value)


def gset(**kwargs):
    cmd("\n".join(parse_set(key, value) for key, value in kwargs.items()))


def term(term, **kwargs):
    size     = kwargs.get("size")
    font     = str(kwargs.get("font", "Verdena"))
    fontsize = float(kwargs.get("fontsize", 12))
    enhanced = bool(kwargs.get("enhanced", False))

    txt = "set terminal %s" % term
    
    if enhanced:
        txt += " enhanced"
    
    if size is not None:
        txt += " size {},{}".format(size[0], size[1])
    
    cmd("%s font '%s,%g'" % (txt, font, fontsize))


def output(outfile, **kwargs):
    term(**kwargs)
    cmd("\nset output '%s'" % outfile)
    

def title(title):
    gset(title=title)


def get_format(dtype, cols=None, shape=None, matrix=False):
    txt = "binary format='%s' endian=big" % formats[dtype.upper()]
    
    if matrix:
        return txt + " matrix"
    elif cols is not None:
        return txt + " record=(%d)" % cols
    elif shape is not None:
        rows, cols = shape.azi(), shape.rng()
        return txt + " array=%dx%d" % (cols, rows)
    else:
        return txt 



def histo(data, dtype, out, limits=None, nbin=10, **kwargs):
    output(out, **kwargs)
    
    if limits is None:
        limits = (-1e6, 1e6)
    
    xlabel = kwargs.pop("xlabel", "x")
    
    gset(ylabel="Count", xlabel=xlabel, **kwargs)
    
    fmt = get_format(dtype)
    
    cmd(
    """
    Min = {}  # where binning starts
    Max = {}  # where binning ends
    n   = {}  # the number of bins
    width = (Max - Min) / n # binwidth
    bin(x) = width * (floor((x - Min) / width) + 0.5) + Min
    set boxwidth width * 0.9
    set style fill solid 0.5  # fill style
    
    plot '{infile}' {fmt} using (bin($1)):(1) \
    smooth freq with boxes lc rgb "red" notitle
    """ .format(limits[0], limits[1], nbin, fmt=fmt, infile=data))
    
    plot(**kwargs)
