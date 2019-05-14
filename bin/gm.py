#!/usr/bin/env python

from functools import partial

from gamma import (display, raster, Argp, display_parser, raster_parser,
                   Processing)


def proc(args):
    proc = Processing(args.conf_file)
    proc.run_steps(args)


def execute(process, args):
    process.parse_args(args)


def _display(args):
    display(args.datfile, **vars(args))


def _raster(args):
    raster(**vars(args))


def main():
    
    narg = Argp.narg

    ap = Argp(subcmd=True)

    # ***********
    # * ts_prep *
    # ***********
    
    ap.subcmd("proc", proc, *Processing.args)

    ap.subcmd("ras", _raster, parents=[raster_parser])
    ap.subcmd("dis", _display, parents=[display_parser])

    args = ap.parse_args()
    args.fun(args)

    
if __name__ == "__main__":
    main()

