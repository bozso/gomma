package gamma;

import (
    "log";
    "os";
    "encoding/json";
);


type (
    general struct {
        Cache_path string `json:"CACHE_PATH,omitempty"`;
        Slc_data, Output_dir, Pol, Metafile string;
        Range_looks, Azimuth_looks int;
    };
    
    preselect struct {
        date_start, date_stop, master_date, lower_left, upper_right string;
        check_zips bool
    };
    
    geocoding struct {
        dem_path string;
        iter, rng_overlap, azi_overlap, npoly int;
        dem_lat_ovs, dem_lon_ovs float64;
    };
    
    coreg struct {
        cc_thresh, fraction_thresh, ph_stdev_thresh float64;
    };
    
    ifg_select struct {
        bperp_min, bperp_max float64;
        delta_t_min, delta_t_max int;
    };
    
    coherence struct {
        weight_type string;
        box_min, box_max, slope_corr_thresh  float64;
        slope_win int;
    };
    
    config struct {
        General general;
        preselect preselect;
        geocoding geocoding;
        coreg coreg;
        ifg_select ifg_select;
        coherence coherence;
    };
);


func ParseConfig(path string) config {
    var ret config;
    err := json.Unmarshal(ReadFile(path), ret);
    Check(err, "Failed to parse json file: %v", path);
    
    return ret;
}


var defaultConfig = config{
        general{
            Cache_path:"/mnt/bozso_i/cache",
            Pol:"vv",
            Range_looks:1,
            Azimuth_looks:1,
        },
        
        preselect{
            master_date: "auto",
            check_zips: false,
        },
        
        geocoding{
            dem_path: "/home/istvan/DEM/srtm.vrt",
            iter:1,
            rng_overlap: 100,
            azi_overlap: 100,
            npoly: 1,
            dem_lat_ovs: 1.0,
            dem_lon_ovs: 1.0,
        },
        
        coreg{
            cc_thresh: 0.8,
            fraction_thresh: 0.01,
            ph_stdev_thresh: 0.8,
        },
        
        ifg_select{
            bperp_min: 0.0,
            bperp_max: 150.0,
            delta_t_min: 0,
            delta_t_max: 15,
        },
        
        coherence{
            weight_type: "gaussian",
            box_min: 3.0,
            box_max: 9.0,
            slope_corr_thresh: 0.4,
            slope_win: 5,
        },
};


func DefaultConfig(path string) {
    log.Println(defaultConfig);
    
    out, err := json.MarshalIndent(defaultConfig,  "", "    ");
    Check(err, "Failed to json encode default configuratio!");
    
    
    f, err := os.Create(path);
    Check(err, "Failed to create file: %v!", path);
    defer f.Close();
    
    _, err = f.Write(out);
    Check(err, "Failed to write to: %v", path);
}





/*
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


[reflector]
# station file containing reflector parameters
station_file = /mnt/Dszekcso/NET/D_160928.stn

# oversempling factor for SLC search
ref_ovs = 16

# size of search window
ref_win = 3
*/