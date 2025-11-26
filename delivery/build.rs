fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile protobuf files using prost
    prost_build::Config::new()
        .out_dir("src/domain")
        .compile_well_known_types()
        .extern_path(".google.protobuf", "::pbjson_types")
        .compile_protos(
            &[
                "domain/common/v1/common.proto",
                "domain/commands/v1/commands.proto",
                "domain/commands/v1/responses.proto",
                "domain/events/v1/events.proto",
            ],
            &["domain"],
        )?;
    
    // Tell Cargo to rerun this build script if proto files change
    println!("cargo:rerun-if-changed=domain/");
    
    Ok(())
}

