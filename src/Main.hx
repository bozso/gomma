import gamma.common;

enum ItemType {
  Key;
  Sword(name:String, attack:Int);
  Shield(name:String, defense:Int);
}


typedef CmdFun = Array<String> -> String;


class Main {
    static function choose(a: ItemType) {
        return switch(a) {
            case ItemType.Key:
                "Key";
            case ItemType.Sword(name, attack):
                "Sword";
            case ItemType.Shield(name, defense):
                "Shield";
        }
    }
    //static function inner(args: Array<String>): String {
    //    return "a";
    //}
    
    static function make_cmd(name: String): CmdFun {
        return function (args: Array<String>): String {
            return "a";
        }
        //return inner;
    }
    
    static function test(a: ItemType) {
        var output = new sys.io.Process("ls", []).stdout.readAll().toString();
        trace(a);
    }
    
    static function main() {
        Main.test(Sword("A", 11));
        trace("Hello World!");
    }
}