from subprocess import Popen, STDOUT, PIPE
from collections import namedtuple

import gamma as gm

__all__ = [
    "gset",
    "histo",
    "cmd",
    "output",
    "title",
    "plot",
    "term",
    "get_format",
    "montage",
    "make_palette",
    "make_colorbar",
    "colors"
]

_convert = gm.make_cmd("convert")
_montage = gm.make_cmd("montage")
gnuplot = gm.make_cmd("gnuplot")


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


RGB = namedtuple("RGB", "r g b")

colors = {
    "black"     : RGB(  0,    0,    0),
    "white"     : RGB(255,  255,  255),
    "red"       : RGB(255,    0,    0),
    "lime"      : RGB(  0,  255,    0),
    "blue"      : RGB(  0,    0,  255),
    "yellow"    : RGB(255,  255,    0),
    "aqua"      : RGB(  0,  255,  255),
    "magenta"   : RGB(255,    0,  255),
    "silver"    : RGB(192,  192,  192),
    "gray"      : RGB(128,  128,  128),
    "maroon"    : RGB(128,    0,    0),
    "olive"     : RGB(128,  128,    0),
    "olive"     : RGB(128,  128,    0),
    "green"     : RGB(  0,  128,    0),
    "purple"    : RGB(128,    0,  128),
    "teal"      : RGB(  0,  128,  128),
    "navy"      : RGB(  0,    0,  128)
}


def parse_opt(key, value):
    if value is True:
        return "-%s" % (key)
    else:
        return "-%s %s" % (key, value)


def montage(out, *args, **kwargs):
    size = kwargs.pop("size")
    debug = bool(kwargs.get("debug", False))
    
    
    if size is not None:
        kwargs["resize"] = "x".join(str(pixel) if pixel is not None else ""
                                    for pixel in size)
    
    options = " ".join(parse_opt(key, value)
                       for key, value in kwargs.items())
    
    
    
    files = " ".join(str(arg) for arg in args)
    
    _montage(files, files, options, out, debug=debug)


def palette_line(line):
    return " ".join(str(float(elem) / 255.0) for elem in line.split())


def make_palette(cmap):
    ret = "defined (%s)"
    
    with open(cmap) as f:
        return ret % ",".join("%d %s" % (ii, palette_line(line))
                              for ii, line in enumerate(f))


def make_colorbar(inras, outras, cmap, title="", ratio=1, start=0.0,
                  stop=255.0):
    
    tmp = Files.get_tmp()
    cbar, script = tmp + ".png", tmp + ".prt"
    cmap = pth.join(gamma_cmaps, "%s" % cmap)
    
    palette = "set palette %s" % make_palette(cmap)
    
    
    with open(script, "w") as f:
        f.write(cbar_tpl.format(out=cbar, xmin=start, xmax=stop,
                                ratio=ratio, title=title, palette=palette))
    
    _gnuplot(script)
    _convert(inras, cbar, "+append", outras)
    
    rm(cbar, script)


cbar_tpl = \
"""\
set terminal pngcairo size 200,800
set output "{out}"
set pm3d map

g(x,y) = y

set yrange [{xmin}:{xmax}]
# set ytics 0.2
set ytics scale 1.5 nomirror
# set mytics 2
set size ratio 1e{ratio}

{palette}

unset colorbox; unset key; set tics out; unset xtics
set title "{title}"
splot g(x,y)
"""

