import nake, strformat, os

const
  goc = "go build"
  src = "gamma.go"
  exe = "gamma"
  mountRoot = joinPath("/home", "istvan", "mount")

proc compileGo(src: string, flags: varargs[string, `$`]) =
  shell(goc, flags.join(" -"), src)

proc deployExe(src, serverName: string) =
  let dst = mountRoot.joinPath(serverName, "packages", "bin", src)
  
  if dst.needsRefresh(src):
    try:
      copyFileWithPermissions(src, dst)
    except OSError:
      discard


task "debug-build", "Compile debug build":
  compileGo(src)

task "release-build", "Compile debug build":
  compileGo(src, "-ldflags=\"-s -w\"")

task "deploy", "Deploy gamma executable to servers":
  runTask("release-build")
  
  deployExe("gamma", "robosztus")
  deployExe("gamma", "zafir")
