tonic::include_proto!("api");

mod tests {
    use prost::Message;
    use std::fs::File;
    use std::io::{self, Write};

    use super::*;
    #[test]
    fn gen_benchmark_body() {
        let body = PsiExecuteRequest {
            header: None,
            keys: vec!["1".into(), "2".into()],
        };

        let mut file = File::create("request_body.bin").unwrap();
        file.write_all(&body.encode_to_vec()).unwrap();
        file.flush().unwrap();
    }
}
