install_name_tool -change /usr/lib/libGLEW.1.9.0.dylib @executable_path/libGLEW.1.9.0.dylib tdos
#install_name_tool -change /usr/lib/libSystem.B.dylib @executable_path/libSystem.B.dylib tdos
install_name_tool -change @executable_path/libglfw.dylib @executable_path/libglfw.dylib tdos
otool -L tdos

install_name_tool -change /usr/lib/libobjc.A.dylib @executable_path/libobjc.A.dylib libglfw.dylib
otool -L libglfw.dylib
