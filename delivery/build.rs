fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile protobuf files using prost
    prost_build::Config::new()
        .out_dir("src/domain")
        .compile_well_known_types()
        .extern_path(".google.protobuf", "::pbjson_types")
        .compile_protos(
            &[
                "src/domain/delivery/common/v1/common.proto",
                "src/domain/delivery/commands/v1/commands.proto",
                "src/domain/delivery/commands/v1/responses.proto",
                "src/domain/delivery/events/v1/events.proto",
                "src/domain/delivery/queries/v1/queries.proto",
            ],
            &["src/domain"],
        )?;
    
    // Tell Cargo to rerun this build script if proto files change
    println!("cargo:rerun-if-changed=src/domain/delivery/");
    
    Ok(())
}

