#!/usr/bin/env python

import numpy as np
from gnuplot import Gnuplot

def main():
    
    mli = np.fromfile("/media/nas1/bozso_i/dszekcso/stations/IB1/20160912.ptr", dtype=">f")
    #lon = np.fromfile("/media/nas1/Kulcs/PS_DSC_INSAR_20170420_preproc/20170420.lon", dtype=">f")
    #lat = np.fromfile("/media/nas1/Kulcs/PS_DSC_INSAR_20170420_preproc/20170420.lat", dtype=">f")
    
    #mli = mli[lon 
    
    hist, edges = np.histogram(mli, bins=int(1e2))
    
    gp = Gnuplot(persist=1)
    gp.plot(gp.histo(edges, np.log10(hist)))
    
if __name__ == "__main__":
    main()
