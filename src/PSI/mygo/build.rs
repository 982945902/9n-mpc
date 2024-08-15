fn main() {
    tonic_build::configure()
        .compile(&["proto/api.proto"], &["proto"])
        .unwrap();
}