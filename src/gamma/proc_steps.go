package gamma;

import (
    "os";
    "encoding/json";
);


type (
    minmax struct {
        Min, Max float64;
    }

    general struct {
        CachePath string `json:"CACHE_PATH,omitempty"`;
        DataPath, OutputDir, Pol, Metafile string;
        RangeLooks, AzimuthLooks int;
    };
    
    preselect struct {
        DateStart, DateStop, MasterDate, LowerLeft, UpperRight string;
        CheckZips bool
    };
    
    geocoding struct {
        DEMPath string;
        Iter, RangeOverlap, AzimuthOverlap, NPoly int;
        DEMLatOversampling, DEMLonOversampling float64;
    };
    
    coreg struct {
        CoherenceThresh, FractionThresh, PhaseStdevThresh float64;
    };
    
    ifgSelect struct {
        Bperp minmax;
        DeltaT minmax;
    };
    
    coherence struct {
        WeightType string;
        Box minmax;
        SlopeCorrelationThresh  float64;
        SlopeWindow int;
    };
    
    config struct {
        General general;
        Preselect preselect;
        Geocoding geocoding;
        Coreg coreg;
        IFGSelect ifgSelect;
        Coherence coherence;
    };
);


func ParseConfig(path string) (config, error) {
    handle := Handler("ParseConfig");
    var ret config;
    
    data, err := ReadFile(path);
    
    if err != nil {
        return config{}, handle(err, "Failed to read file:  '%s'!", path);
    }
    
    if err := json.Unmarshal(data, ret); err != nil {
        return config{}, handle(err, "Failed to parse json data: %s'!", data);
    }
    
    return ret, nil;
}


var defaultConfig = config{
        General: general{
            CachePath:"/mnt/bozso_i/cache",
            Pol:"vv",
            RangeLooks:1,
            AzimuthLooks:1,
        },
        
        Preselect: preselect{
            MasterDate: "auto",
            CheckZips: false,
        },
        
        Geocoding: geocoding{
            DEMPath: "/home/istvan/DEM/srtm.vrt",
            Iter:1,
            RangeOverlap: 100,
            AzimuthOverlap: 100,
            NPoly: 1,
            DEMLatOversampling: 1.0,
            DEMLonOversampling: 1.0,
        },
        
        Coreg: coreg{
            CoherenceThresh: 0.8,
            FractionThresh: 0.01,
            PhaseStdevThresh: 0.8,
        },
        
        IFGSelect: ifgSelect{
            Bperp: minmax{Min: 0.0, Max: 150.0},
            DeltaT: minmax{Min: 0.0, Max: 15.0},
        },
        
        Coherence: coherence{
            WeightType: "gaussian",
            Box: minmax{Min: 3.0, Max: 9.0},
            SlopeCorrelationThresh: 0.4,
            SlopeWindow: 5,
        },
};


func MakeDefaultConfig(path string) error {
    handle := Handler("MakeDefaultConfig");
    
    out, err := json.MarshalIndent(defaultConfig,  "", "    ");
    
    if err != nil {
        return handle(err, "Failed to json encode default configuration!");
    }
    
    
    
    f, err := os.Create(path);
    if err != nil {
        return handle(err, "Failed to create file: %v!", path);
    }
    defer f.Close();
    
    _, err = f.Write(out);
    
    if err != nil {
        return handle(err, "Failed to write to file '%v'!", path);
    }
    
    return nil;
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