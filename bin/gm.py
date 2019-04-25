#!/usr/bin/env python

from functools import partial

import gamma.processing_steps as gs
from gamma.processing_steps import pk_load

from gamma.base import display, raster, Argp, display_parser, raster_parser


TSProc = gs.Processing(
    gs.make_step("select_bursts", gs.select_bursts),
    gs.make_step("import",        gs.import_slc),
    gs.make_step("merge",         gs.merge_slcs),
    gs.make_step("quick_mli",     gs.quicklook_mli, False),
    gs.make_step("mosaic",        gs.mosaic_tops, False),
    gs.make_step("iono",          gs.check_ionoshpere, False),
    gs.make_step("geocode",       gs.geocode_master),
    gs.make_step("geo_check",     gs.check_geocode, False),
    gs.make_step("coreg",         gs.coreg_slcs),
)


IPTA = gs.Processing(
    gs.make_step("deramp", gs.deramp),
    gs.make_step("base_plot", gs.base_plot),
    gs.make_step("avg_mli", gs.avg_mli),
)


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
    
    ap.subcmd("tsprep", partial(execute, TSProc), *TSProc.add_args())
    ap.subcmd("ipta", partial(execute, IPTA),
              *IPTA.add_args("ipta_conf.py", "ipta_proc.log"))

    ap.subcmd("ras", _raster, parents=[raster_parser])
    ap.subcmd("dis", _display, parents=[display_parser])

    args = ap.parse_args()
    args.fun(args)

    
if __name__ == "__main__":
    main()

