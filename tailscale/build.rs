use std::{env, path::PathBuf, process::Command};

fn main() {
    println!("cargo:rerun-if-changed=libtailscale");
    let outdir = env::var("OUT_DIR").unwrap();
    let out = format!("{outdir}/libtailscale.a");
    let status = Command::new("go")
        .args([
            "build",
            "-C",
            "libtailscale",
            "-buildmode=c-archive",
            "-o",
            &out,
        ])
        .status()
        .expect("can't build go library");

    assert!(status.success(), "failed to build go library");

    println!("cargo:rustc-link-search={outdir}");
    println!("cargo:rustc-link-lib=static=tailscale");

    #[cfg(target_os = "macos")]
    {
        println!("cargo:rustc-flags=-l framework=CoreFoundation -l framework=Security");
    }

    let bindings = bindgen::Builder::default()
        .header("libtailscale/tailscale.h")
        // Tell cargo to invalidate the built crate whenever any of the
        // included header files changed.
        .parse_callbacks(Box::new(bindgen::CargoCallbacks::new()))
        .generate()
        .expect("Unable to generate bindings");

    // Write the bindings to the $OUT_DIR/libtailscale.rs file.
    let out_path = PathBuf::from(env::var("OUT_DIR").unwrap());
    bindings
        .write_to_file(out_path.join("libtailscale.rs"))
        .expect("Couldn't write bindings!");
}
