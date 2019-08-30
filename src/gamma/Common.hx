package gamma;


import python.lib.Glob;
import python.lib.os.Path.join;
import python.lib.datetime.Datetime;
import python.Exceptions.RuntimeError;

import sys.io.Process;
import Std;


typedef CmdFun = Array<Printable> -> Bool -> String;

private class Settings {
    public var ras_ext: String;
    public var path: String;
    public var modules: Array<String>;
    public var libpaths: String;
    public var templates: Templates;
    public var cache_default_path: String;
    
    public function new(ras_ext, path, modules, libpaths, templates,
                 cache_default_path)
    {
        this.ras_ext = ras_ext;
        this.path = path;
        this.modules = modules;
        this.libpaths = libpaths;
        this.templates = templates;
        this.cache_default_path = cache_default_path;
    }
}

private class Templates {
    public var IW: String;
    public var short: String;
    public var long: String;
    public var tab: String;
    
    public function new(IW, short, long, tab)
    {
        this.IW = IW;
        this.short = short;
        this.long = long;
        this.tab = tab;
    }
}


interface Printable {
    public function toString(): String;
}


enum DateFormat {
    Short;
    Long;
}

class Common {
    public static var versions = [
        "20181130" => "/home/istvan/progs/GAMMA_SOFTWARE-20181130"
    ];

    public static var settings = new Settings(
        "bmp",
        versions["20181130"],
        ["DIFF", "DISP", "ISP", "LAT", "IPTA"],
        "/home/istvan/miniconda3/lib:",
        new Templates(
            "{date}_iw{iw}.{pol}.slc",
            "%Y%m%d",
            "%Y%m%dT%H%M%S",
            "{date}.{pol}.SLC_tab"
        ),
        "/mnt/bozso_i/cache");
    
    private static var  gamma_commands = [
        for (module in Common.settings.modules)
        for (path in ["bin", "scripts"])
        for (binfile in Glob.iglob(join(settings.path, module, path, "*")))
        binfile
    ];
    
    private static function make_cmd(name: String): CmdFun {
        return function (args: Array<Printable>, debug: Bool = false): String {
            var args = [for (elem in args) args.toString()];
            var cmd = "${name} ${args}";
            
            if (debug) {
                trace("Command: ${cmd}");
                return "";
            }
            
            var proc = new Process("ls", args);
            
            if (proc.exitCode() != 0) {
                throw new RuntimeError("\nNon zero returncode from command: 
                                       \n'${cmd}'\n \nOUTPUT OF THE COMMAND: 
                                       \n\n{proc.stderr.readAll()}");
            }
            
            return proc.stdout.readAll().toString();
        }
    }
}


class Date {
    public var start: Datetime;
    public var stop: Datetime;
    public var center: Datetime;
    
    public function new(start, stop, center) {
        this.start = start;
        this.stop = stop;
        this.center = center;
    }
    
    
    static public function parse_short(str: String): Datetime {
        return new Datetime(parseInt(str.substr(0, 4)),
                            parseInt(str.substr(4, 2)),
                            parseInt(str.substr(6)));
    }
    
    
    static public function parse_long(str: String): Datetime {
        return new Datetime(parseInt(str.substr(0, 4)),
                            parseInt(str.substr(4, 2)),
                            parseInt(str.substr(6)));
    }
    
    
    static public function from_string(sstart: String, sstop: String, 
                                       format: DateFormat): Date {
        var start: Datetime;
        var stop: Datetime;
        
        switch(format) {
            case DateFormat.Short:
                start = Date.parse_short(sstart);
                stop = Date.parse_short(sstop);
            case DateFormat.Long:
                start = Date.parse_long(sstart);
                stop = Date.parse_long(sstop);
        }
        
        var center = (start - stop) / 2.0;
        center += stop;
        
        return new Date(start, stop, center);
        
    }
}
    
    


class Point {
    public var x: Float;
    public var y: Float;
    
    public function new(x, y) {
        this.x = x;
        this.y = y;
    }
    
    public function in_rect(rect: Rect) {
        return (this.x > rect.min.x && this.x < rect.max.x &&
                this.y > rect.min.y && this.y < rect.max.y);
    }
}


class Rect {
    public var max: Point;
    public var min: Point;
    
    public function new(min, max) {
        this.max = max;
        this.min = min;
    }
}

