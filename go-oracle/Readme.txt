For build:

1. Put oci8.pc path to your PKG_CONFIG_PATH environment variable (modify it to match your client / database include paths)
2. You need:
   - go
   - oracle client, instant client or database 12.x (x64 !!!)
   - gcc from mingw64 - mine is installed in C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     and I put it in my path
     choose latest version (default), architecture x86_64, threads win32, exception seh
   - pkg-config for Windows
     - copy pkg-config_0.26-1_win32.zip/bin/pkg-config.exe into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     - copy gettext-runtime_0.18.1.1-2_win32.zip/bin/intl.dll into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     - copy glib_2.28.8-1_win32.zip/bin/libglib-2.0-0.dll into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
3. go get github.com/mattn/go-oci8
