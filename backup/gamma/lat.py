import gamma as gm

from logging import getLogger

log = getLogger("gamma.lat")

gp = gm.gp


__all__ = ("single_class_mapping", "radcal_MLI")


def single_class_mapping(*args, width=None, start=1, nlines=0, avg_rng=None,
                         avg_azi=None, flip=False, ras=None):
    lr = -1 if flip else 1
    
    narg = len(args)
    
    assert narg % 3, "Number of arguments must be divisable by 3."
    assert ras is not None, "ras has to be defined"
    assert width is not None, "width has to be defined"
    
    for item in args[::3]:
        print(item)
    
    exit()
    
    #rngs = tuple(item.rng() for item in args[::3])
    #azis = tuple(item.azi() for item in args[::3])
    
    #assert(gb.all_same(rngs), "Width of input files is not all the same")
    #assert(gb.all_same(azis), "Lines of input files is not all the same")
    
    gm.single_class_mapping(narg, *args, ras, width, start, nlines,
                            arng, aazi, lr)


def radcal_MLI(params):

    output_dir, master_date = get_out_master(params)
    
    mli = pth.join(output_dir, "coreg_out", master_date + ".rmli")
    mli_par = mli + ".par"

    mli_width = get_rng(mli_par)

    avg_rng, avg_azi = avg_factor(mli_par)

    for rmli_s0 in iglob(pth.join(output_dir, "coreg_out", "*.rmli.s0")):
        gm.raspwr(rmli_s0, mli_width, None, None, avg_rng, avg_azi, None,
                  None, None, rmli_s0 + ".bmp")

    mli = pth.join(output_dir, coreg_out, master_date + ".rmli")
    mli_par = mli + ".par"

    mli_width = get_rng(mli_par)

    geo_path = pth.join(output_dir, "geo", master_date)

    pix_sigma0 = geo_path + ".pix_sigma0"
    pix_ell = geo_path + ".pix_ell"
    pix_rdc = geo_path + ".pix_rdc"

    gm.radcal_MLI(mli, mli_par, None, "nnn", None, 0, 0, 1, 0, 0, pix_ell)
    gm.ratio(pix_ell, pix_sigma0, pix_rdc, mli_width, 1, 1, 0)

    for rmli in iglob(pth.join(output_dir, "coreg_out", "*.rmli")):
        gm.product(rmli, pix_rdc, rmli + ".s0", mli_width, 1, 1, 0)
