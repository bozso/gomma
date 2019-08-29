#!/usr/bin/env python

from functools import partial

from gamma import *
from utils import *

class GammaCmd(CParse):
    commands = ("display", "raster", "proc")
    
    def __init__(self, **kwargs):
        CParse.__init__(self, **kwargs)
    
    @annot(
        datfile=pos(help="Datafile."),
        mode=opt(help="Command to use."),
        flip=flag(help="Flip image left-right."),
        rng=opt(help="Range samples."),
        parfile=opt(help="Parameter file."),
        image_format=opt(help="Image format."),
        debug=flag(help="Debug mode.")
    )
    def display(self):
        args = self.args
        display(args.datfile, **vars(args))
    
    
    @annot(
        datfile=pos(help="Datafile."),
        mode=opt(help="Command to use."),
        flip=flag(help="Flip image left-right."),
        rng=opt(help="Range samples."),
        parfile=opt(help="Parameter file."),
        image_format=opt(help="Image format."),
        debug=flag(help="Debug mode."),
        raster=opt(help="Output raster file."),
        azi=opt(help="Azimuth lines."),
        avg_fact=opt(help="Pixel averaging factor.", type=int, default=750)
    )
    def raster(self):
        raster(**vars(self.args))
    
    @annot(
        conf_file=pos(type=str, help="File holding information about "
                      "processing steps."),
        
        step=opt(type=str, help="Single processing step to be executed"),
        
        start=opt(type=str, help="Starting processing step. Processing steps "
                  "will be executed until processing step defined by "
                  "--stop is reached."),
                  
        stop=opt(type=str, help="Last processing step to be executed."),
        
        logfile=opt(type=str, help="Log messages will be saved here."),
        
        loglevel=opt(type=str, help="Level of logging.",
                     choices=["info", "debug", "error"], default="info"),
        
        skip_optional=flag(help="If set the processing will skip optional steps."),
        
        show_steps=flag(help="If set, just print the processing steps."),
        
        info=flag(help="Dumps information about the processing to the terminal.")
    )
    def proc(self):
        Processing(self.args).run_steps()
        

def main():
    gm.make_cmd(5.0)
    
    GammaCmd().parse().args.fun()
    
    
    
if __name__ == "__main__":
    main()

