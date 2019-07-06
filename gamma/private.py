from datetime import datetime
from subprocess import check_output, CalledProcessError, STDOUT
from logging import getLogger
from collections import Iterable
from six import string_types
from re import match
from shlex import split
from math import sqrt
from glob import iglob
from os.path import join as pjoin, splitext

log = getLogger("gamma.private")

r_tiff_tpl = ".*.SAFE/measurement/s1.*-iw{iw}-slc-{pol}.*.tiff"
r_annot_tpl = ".*.SAFE/annotation/s1.*-iw{iw}-slc-{pol}.*.xml"
r_calib_tpl = ".*.SAFE/annotation/calibration/calibration"\
              "-s1.*-iw{iw}-slc-{pol}.*.xml"
r_noise_tpl = ".*.SAFE/annotation/calibration/noise-s1.*-"\
              "iw{iw}-slc-{pol}.*.xml"


ScanSAR = True


def _diff_burst(burst1, burst2):
    
    diff_sqrt = sqrt((burst1 - burst2) * (burst1 - burst2))
    
    return int(burst1 - burst2 + 1.0
               + ((burst1 - burst2) / (0.001 + diff_sqrt)) * 0.5)


def burst_selection_helper(ref_burst, slc_burst):
    if ref_burst is not None:
        iw_start_burst = slc_burst[0]
    
        diff = [_diff_burst(ref_burst[0], iw_start_burst),
                _diff_burst(ref_burst[-1], iw_start_burst)]
        
        total_slc_bursts = len(slc_burst)
    
        if diff[1] < 1 or diff[0] > total_slc_bursts:
            return None
    
        if diff[0] <= 0:
            diff[0] = 1
    
        if diff[1] > total_slc_bursts:
            diff[1] = total_slc_bursts
    
        return tuple((diff[0], diff[1]))
    else:
        return None


def extract_file(slc_zip, regex, out_path):
    return tuple(slc_zip.extract(elem, out_path)
                 for elem in slc_zip.namelist() if match(regex, elem))


def iterable(arg):
    return (isinstance(arg, Iterable)
            and not isinstance(arg, string_types))






def cmd(command, *args, **kwargs):
    debug = kwargs.pop("debug", False)
    
    Cmd = "%s %s" % (command, " ".join(_proc_arg(arg) for arg in args))
    
    log.debug('Issued command is "%s"' % Cmd)
    
    if debug:
        print(Cmd)
        return
    
    try:
        proc = check_output(split(Cmd), stderr=STDOUT)
    except CalledProcessError as e:
        log.error("\nNon zero returncode from command: \n'{}'\n"
                  "\nOUTPUT OF THE COMMAND: \n\n{}\nRETURNCODE was: {}"
                  .format(Cmd, e.output.decode(), e.returncode))

        raise e

    return proc




"""This is a backport of shutil.get_terminal_size from Python 3.3.

The original implementation is in C, but here we use the ctypes and
fcntl modules to create a pure Python version of os.get_terminal_size.
"""

import os
import struct
import sys

from collections import namedtuple

__all__ = ["get_terminal_size"]


terminal_size = namedtuple("terminal_size", "columns lines")

try:
    from ctypes import windll, create_string_buffer

    _handles = {
        0: windll.kernel32.GetStdHandle(-10),
        1: windll.kernel32.GetStdHandle(-11),
        2: windll.kernel32.GetStdHandle(-12),
    }

    def _get_terminal_size(fd):
        columns = lines = 0

        try:
            handle = _handles[fd]
            csbi = create_string_buffer(22)
            res = windll.kernel32.GetConsoleScreenBufferInfo(handle, csbi)
            if res:
                res = struct.unpack("hhhhHhhhhhh", csbi.raw)
                left, top, right, bottom = res[5:9]
                columns = right - left + 1
                lines = bottom - top + 1
        except Exception:
            pass

        return terminal_size(columns, lines)

except ImportError:
    import fcntl
    import termios

    def _get_terminal_size(fd):
        try:
            res = fcntl.ioctl(fd, termios.TIOCGWINSZ, b"\x00" * 4)
            lines, columns = struct.unpack("hh", res)
        except Exception:
            columns = lines = 0

        return terminal_size(columns, lines)


def get_terminal_size(fallback=(80, 24)):
    """Get the size of the terminal window.

    For each of the two dimensions, the environment variable, COLUMNS
    and LINES respectively, is checked. If the variable is defined and
    the value is a positive integer, it is used.

    When COLUMNS or LINES is not defined, which is the common case,
    the terminal connected to sys.__stdout__ is queried
    by invoking os.get_terminal_size.

    If the terminal size cannot be successfully queried, either because
    the system doesn't support querying, or because we are not
    connected to a terminal, the value given in fallback parameter
    is used. Fallback defaults to (80, 24) which is the default
    size used by many terminal emulators.

    The value returned is a named tuple of type os.terminal_size.
    """
    # Try the environment first
    try:
        columns = int(os.environ["COLUMNS"])
    except (KeyError, ValueError):
        columns = 0

    try:
        lines = int(os.environ["LINES"])
    except (KeyError, ValueError):
        lines = 0

    # Only query if necessary
    if columns <= 0 or lines <= 0:
        try:
            size = _get_terminal_size(sys.__stdout__.fileno())
        except (NameError, OSError):
            size = terminal_size(*fallback)

        if columns <= 0:
            columns = size.columns
        if lines <= 0:
            lines = size.lines

    return terminal_size(columns, lines)


def avg_factor(rng, azi, comp_fact=750):
    
    avg_rng = int(float(rng) / comp_fact)

    if avg_rng < 1:
        avg_rng = 1

    avg_azi = int(float(azi) / comp_fact)

    if avg_azi < 1:
        avg_azi = 1

    return avg_rng, avg_azi


inits = {
    "tsprep": \
    """\
    [general]
    slc_data = 
    output_dir = .
    pol = vv
    date_start = 
    date_stop = 
    master_date = 
    iw1 = 
    iw2 = 
    iw3 = 
    check_zips = False
    range_looks = 1
    azimuth_looks = 5
    metafile = processing.pkl
    
    [check_ionosphere]
    # range and azimuth window size used in offset estimation
    rng_win = 256
    azi_win = 256
    
    # threshold value used in offset estimation
    iono_thresh = 0.1
    
    # range and azimuth step used in offset estimation, 
    # default (rng|azi)_win / 4
    rng_step = 
    azi_step = 
    
    [geocoding]
    # Path to DEM vrt file. Only vrt files accepted
    dem_path = /home/istvan/DEM/srtm.vrt
    
    # number of offset refinement iterations
    iter = 1
    
    # DEM oversampling factor
    dem_lat_ovs = 1.0
    dem_lon_ovs = 1.0
    
    # Number of offset estimate windows in range and azimuth
    n_rng_off = 32
    n_azi_off = 32
    
    # Overlap of windows in pixels
    rng_overlap = 100
    azi_overlap = 100
    
    # number of polynom coefficients. Availble values: 1, 3, 4, 6
    npoly = 1
    
    [coreg]
    # coherence threshold used
    cc_thresh = 0.8
    
    # minimum valid fraction of unwrapped phase values
    fraction_thresh = 0.01
    
    # phase standard deviation threshold
    ph_stdev_thresh = 0.8

    range_looks = 10
    azimuth_looks = 2

    
    [ifg_select]
    # min. and max. perpendicular baseline in meters
    bperp_min = 0.0
    bperp_max = 150.0
    
    # min. and max. temporal baseline in days
    delta_t_min = 0
    delta_t_max = 15
    
    [coherence]
    # estimation box min., max. sizes
    box_min = 3.0
    box_max = 9.0
    
    # phase slope estimation win size and correlation threshold 
    # for accepting phase slope estimates
    slope_win = 5
    slope_corr_thresh = 0.4
    
    # window type, constant or gaussian
    weight_type = gaussian
    
    [reflector]
    # station file containing reflector parameters
    station_file = /mnt/Dszekcso/NET/D_160928.stn
    
    # oversempling factor for SLC search
    ref_ovs = 16
    
    # size of search window
    ref_win = 3
    """
}    
