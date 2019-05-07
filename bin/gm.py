#!/usr/bin/env python

from functools import partial

import gamma.processing_steps as gs
from gamma.processing_steps import pk_load

from gamma.base import display, raster, Argp, display_parser, raster_parser


def proc(args):
    proc = Processing(args.paramfile)
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

