import numpy as np
import re
import subprocess
import matplotlib.pyplot as plt
import matplotlib as mpl
import os.path as pth

from datetime import datetime as dt
from collections import namedtuple

try:
    from urllib2 import urlopen, URLError
except ImportError:
    from urllib.request import urlopen
    from urllib.error import URLError

import gamma.base as gb
import gamma.scripts as gs
import gamma.private as gp

gm = gb.gm

#import cartopy.crs as ccrs
#from mpl_toolkits.axes_grid1 import make_axes_locatable
#import matplotlib.pyplot as plt

Epoch = namedtuple("TECEpoch", ["epoch", "data"])

def download(url, outfile, timeout=5, attempts=3):
    
    _attempts = 0
    
    while _attempts < attempts:
        try:
            response = urlopen(url, timeout=5)
            content = response.read()

            with open(outfile, "wb") as f:
                f.write(content)

            break
        except URLError as e:
            _attempts += 1
            print(type(e))


def parse_map(tecmap, exponent = -1):
    epoch = re.split("EPOCH OF CURRENT MAP", tecmap)[0]
    epoch = ".".join(pad(int(elem)) for elem in epoch.split())
    tecmap = re.split('.*END OF TEC MAP', tecmap)[0]
    
    return (epoch, np.stack((np.fromstring(l, sep=' ')
            for l in re.split('.*LAT/LON1/LON2/DLON/H\\n',tecmap)[1:]))*10**exponent)

    
def get_tecmaps(filename):
    with open(filename) as f:
        ionex = f.read()
        return dict(parse_map(t) for t in ionex.split('START OF TEC MAP')[1:])


def get_tec(tecmap, lat, lon):
    i = round((87.5 - lat)*(tecmap.shape[0]-1)/(2*87.5))
    j = round((180 + lon)*(tecmap.shape[1]-1)/360)
    return tecmap[i,j]


def ionex_filename(year, day, centre, zipped=True):
    return '{}g{:03d}0.{:02d}i{}'.format(centre, day, year % 100, '.Z' if zipped else '')


def ionex_ftp_path(year, day, centre):
    return 'ftp://cddis.gsfc.nasa.gov/gnss/products/ionex/{:04d}/{:03d}/{}'.format(year, day, ionex_filename(year, day, centre))


def ionex_local_path(year, day, centre = 'esa', directory = '/tmp', zipped = False):
    return directory + '/' + ionex_filename(year, day, centre, zipped)

    
def download_ionex(year, day, centre = 'esa', output_dir = '/tmp'):
    outfile = ionex_local_path(year, day, centre, output_dir, zipped=True)
    
    download(ionex_ftp_path(year, day, centre), outfile)
    
    subprocess.call(['gzip', '-d', outfile])


def plot_tec_map(tecmap):
    proj = ccrs.PlateCarree()
    f, ax = plt.subplots(1, 1, subplot_kw=dict(projection=proj))
    ax.coastlines()
    h = plt.imshow(tecmap, cmap='viridis', vmin=0, vmax=100, extent = (-180, 180, -87.5, 87.5), transform=proj)
    plt.title('VTEC map')
    divider = make_axes_locatable(ax)
    ax_cb = divider.new_horizontal(size='5%', pad=0.1, axes_class=plt.Axes)
    f.add_axes(ax_cb)
    cb = plt.colorbar(h, cax=ax_cb)
    plt.rc('text', usetex=True)
    cb.set_label('TECU ($10^{16} \\mathrm{el}/\\mathrm{m}^2$)')


def pad(num):
    if num < 10:
        return "0%d" % num
    else:
        return "%d" % num


def conv(year, month, day):
    return int(dt.strptime("%d.%s.%s" % (year, pad(month), pad(day)), "%Y.%m.%d").strftime("%j"))

def sec2hms(time):
    day = time // (24 * 3600)
    time = time % (24 * 3600)
    hour = time // 3600
    time %= 3600
    minutes = time // 60
    time %= 60
    seconds = time
    
    return "%s.%s.%s" % (pad(hour), pad(minutes), pad(seconds))

def rms(array, dim=-1):
    return np.sqrt(np.sum(array**2, dim) / array.shape[dim])


def parse_parfile(parfile, sep=":"):
    with open(parfile, "r") as f:
        return dict((line.split(sep)[0], line.split(sep)[1].strip())
                     for line in f if sep in line)


def map_extent(pars):
    clon, clat = float(pars["corner_lon"].split()[0]), float(pars["corner_lat"].split()[0])
    plon, plat = float(pars["post_lon"].split()[0]), float(pars["post_lat"].split()[0])
    
    xm = clon + int(pars["width"]) * plon
    y0 = clat + int(pars["nlines"]) * plat
    
    return (clon, xm, y0, clat)


def plot_parajd():
    root = "/media/nas1/szucs_e/Parajd/R7/ipta"
    
    pt = gb.Base(pth.join(root, ""), pt="pt", mask="pmask_B",
                 spar="20180729.rslc.par", pdef="pdef_B",
                 mpar="20180729.rmli.par", ave="ave.rmli",
                 itab="itab_ts", ts="pdisp_ts_B", hgt="phgt2",
                 sigma="psigma_B", herr="pdh_err_B" , deferr="pdef_err_B",
                 ras="pdef_B.ras")
    
    gm.vu_disp(pt.pt, pt.mask, pt.spar, pt.itab, pt.ts, pt.pdef, pt.hgt, pt.sigma, pt.herr, pt.deferr, None, pt.ras, -0.25, 0.0, 2, 128, debug=True)
    return
    
    #gs.prasdt_pwr24(pt.pt, pt.mask, pt.spar, pt.pdef, pt.mpar, pt.ave,
                    #0.1, radius=6.0, imode=3)
    
    gp.cmd("convert", "pdef_B.01.bmp colormap_parajd.png "
            "-resize x1000 +append -pointsize 30 -draw \"text 1200,600 rotate 90 'cm/yr' \" parajd.png")

def plot_iono():

    font = {'family' : 'monospace',
            'weight' : 'medium',
            'size'   : 15}
    
    mpl.rc('font', **font)

    extent = map_extent(parse_parfile("EQA.20090328.dem_par"))
    
    tecu = np.fromfile("20090328_20090628.diff.ion.tecu.geo", dtype=">f").reshape((2668, 2340))
    #tecu = np.fromfile("20090328_20090628.diff.ion.tecu", dtype=">f").reshape((2951, 2336))
    
    inc = np.fromfile("inc", dtype=">f").reshape((2668, 2340))
    #inc = np.fromfile("inc.rdc", dtype=">f").reshape((2951, 2336))
    
    dem = np.fromfile("EQA.20090328.dem", dtype=">f").reshape((2668, 2340))
    
    datas = (tecu, inc, dem)
    titles = ("TEC change", "Inclination")
    cmaps = ("viridis", "viridis", "Greys_r")
    
    f, axes = plt.subplots(1, 3, sharey=True, figsize=(22.5, 7.5))
    
    for data, title, ax, cm in zip(datas, titles, axes, cmaps):
        ax.set_title(title)
        im = ax.imshow(data, cmap=cm)
        f.colorbar(im, ax=ax)
        #ax.set_xlabel('WGS 84 Longitude [deg]')
        ax.set_xlabel('Azimuth')
    
    #axes[0].set_ylabel('WGS 84 Latitude [deg]')
    axes[0].set_ylabel('Range')
    
    f.savefig("maps.png")


def main():
    
    plot_parajd()
    return
    #plot_iono()
    
    extent = list(map_extent(parse_parfile("EQA.20090328.dem_par")))
    
    day1, day2 = conv(2009, 3, 28), conv(2009, 6, 28)
    
    insar1, insar2 = sec2hms(6237.680354), sec2hms(6278.929686)
    
    if 0:
        download_ionex(2009, day1)
        download_ionex(2009, day2)
        return
        
    tecmap1 = get_tecmaps(ionex_local_path(2009, day1))["2009.03.28.02.00.00"]
    tecmap2 = get_tecmaps(ionex_local_path(2009, day2))["2009.06.28.02.00.00"]
    
    dtec = tecmap1 - tecmap2
    
    f = plt.figure()
    plt.imshow(dtec, cmap='viridis', extent=(-180,180,-87.5,87.5))
    plt.axis(extent)
    f.savefig("dtec.png")


if __name__ == "__main__":
    main()
