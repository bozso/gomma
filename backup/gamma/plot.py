from subprocess import Popen, STDOUT, PIPE
from collections import namedtuple

import gamma as gm

__all__ = (
    "montage",
    "make_palette",
    "make_colorbar",
    "colors"
)

_convert = gm.make_cmd("convert")
_montage = gm.make_cmd("montage")


formats = {
    "FCOMPLEX" : "%float32%float32",
    "SCOMPLEX" : "%short%short",
    "FLOAT"    : "%float32",
    "INT"      : "%int",
    "SHORT"    : "%short",
    "BYTE"     : "%byte"
}


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

