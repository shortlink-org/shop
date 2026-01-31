fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile protobuf files using prost (domain messages)
    prost_build::Config::new()
        .out_dir("src/domain/model")
        .compile_well_known_types()
        .extern_path(".google.protobuf", "::pbjson_types")
        .compile_protos(
            &[
                "src/domain/model/delivery/common/v1/common.proto",
                "src/domain/model/delivery/commands/v1/commands.proto",
                "src/domain/model/delivery/commands/v1/responses.proto",
                "src/domain/model/delivery/events/v1/events.proto",
                "src/domain/model/delivery/queries/v1/queries.proto",
            ],
            &["src/domain/model"],
        )?;

    // Compile gRPC service using tonic-prost-build
    tonic_prost_build::configure()
        .build_server(true)
        .build_client(false)
        .compile_protos(
            &["src/infrastructure/rpc/delivery.proto"],
            &["src/infrastructure/rpc"],
        )?;

    // Tell Cargo to rerun this build script if proto files change
    println!("cargo:rerun-if-changed=src/domain/model/delivery/");
    println!("cargo:rerun-if-changed=src/infrastructure/rpc/delivery.proto");

    Ok(())
}

