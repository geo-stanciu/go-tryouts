This is a demo.

Needs:
- golang: https://golang.org

As a first step after cloning this repository, you might need to run the following command:

go get -d

This downloads the needed dependencies.

---------------------------------------------------

Uses the following sql pachages:
- "github.com/denisenkom/go-mssqldb"
- "github.com/go-sql-driver/mysql"
- "github.com/lib/pq"
- "github.com/mattn/go-oci8"

If support is not needed for all of the above databases, remove some of the above imported packages.

---------------------------------------------------

For Oracle driver:

0. Only usable for Oracle 12.1 for now.
1. Put oci8.pc path to your PKG_CONFIG_PATH environment variable.
2. You need:
   - go
   - oracle client or database
   - gcc from mingw64 - mine is installed in C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     and I put it in my path
   - pkg-config for Windows
     - copy pkg-config_0.26-1_win32.zip/bin/pkg-config.exe into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     - copy gettext-runtime_0.18.1.1-2_win32.zip/bin/intl.dll into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
     - copy glib_2.28.8-1_win32.zip/bin/libglib-2.0-0.dll into
        C:\Program Files\mingw-w64\x86_64-6.3.0-win32-seh-rt_v5-rev1\mingw64\bin
3. go get github.com/mattn/go-oci8
