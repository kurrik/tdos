install_name_tool -change /usr/lib/libGLEW.1.9.0.dylib @executable_path/libGLEW.1.9.0.dylib tdos
#install_name_tool -change @executable_path/libglfw.dylib @executable_path/libglfw.dylib tdos
otool -L tdos
