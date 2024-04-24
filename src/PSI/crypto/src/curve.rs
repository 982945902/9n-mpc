// Copyright 2020 The 9nFL Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use arrow::array::{cast, make_array, Array, ArrayData, BinaryArray, BinaryBuilder};
use arrow::pyarrow::PyArrowType;
use curve25519_dalek::montgomery::MontgomeryPoint;
use curve25519_dalek::scalar::Scalar;
use pyo3::exceptions::PyValueError;
use pyo3::prelude::*;
use rand::rngs::StdRng;
use rand::{RngCore, SeedableRng};
use std::default;
use std::sync::Arc;

#[pyclass(module = "crypto.curve", name = "Curve")]
pub(crate) struct Secret(Scalar);

impl Secret {
    fn encrypt_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
        build.append_value(
            (MontgomeryPoint(data.try_into().expect("encrypt accpet 32 bytes pubkey")) * self.0)
                .as_bytes(),
        );
    }

    fn diffie_hellman_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
        build.append_value(
            (MontgomeryPoint(
                data.try_into()
                    .expect("diffie_hellman accpet 32 bytes pubkey"),
            ) * self.0)
                .as_bytes(),
        );
    }

    fn run_impl(
        &self,
        func: fn(s: &Secret, data: &[u8], build: &mut BinaryBuilder),
        array: PyArrowType<ArrayData>,
    ) -> PyResult<PyArrowType<ArrayData>> {
        let array: ArrayData = array.0;
        let array: Arc<dyn Array> = make_array(array);
        let array: &BinaryArray = array
            .as_any()
            .downcast_ref()
            .ok_or_else(|| PyValueError::new_err("expected binary array"))?;

        let mut build = BinaryBuilder::new();
        for i in array.iter() {
            func(self, i.unwrap(), &mut build);
        }

        Ok(PyArrowType(build.finish().into_data()))
    }
}

#[pymethods]
impl Secret {
    #[new]
    #[pyo3(signature = (typ,key))]
    fn pynew(typ: &str, key: Option<[u8; 32]>) -> PyResult<Self> {
        if !SUPPORT_CURVE.contains(&typ) {
            panic!("no support curve type {}", typ)
        }

        Ok(Self(Scalar::from_bytes_mod_order(key.unwrap_or_else(
            || {
                let mut bytes: [u8; 32] = [0; 32];
                StdRng::from_entropy().fill_bytes(&mut bytes);
                bytes
            },
        ))))
    }

    #[pyo3(text_signature = "($self, array)")]
    fn encrypt(&self, array: PyArrowType<ArrayData>) -> PyResult<PyArrowType<ArrayData>> {
        self.run_impl(Secret::encrypt_impl, array)
    }

    #[pyo3(text_signature = "($self, array)")]
    fn diffie_hellman(&self, array: PyArrowType<ArrayData>) -> PyResult<PyArrowType<ArrayData>> {
        self.run_impl(Secret::diffie_hellman_impl, array)
    }
}

#[pyfunction]
pub(crate) fn hash_to_curve(
    typ: &str,
    array: PyArrowType<ArrayData>,
) -> PyResult<PyArrowType<ArrayData>> {
    if !SUPPORT_HASHTOCURVE.contains(&typ) {
        panic!("no support hash_to_curve type {}", typ)
    }

    Ok(array)
}

#[pyfunction]
pub(crate) fn point_octet_marshal(
    typ: &str,
    array: PyArrowType<ArrayData>,
) -> PyResult<PyArrowType<ArrayData>> {
    if !SUPPORT_CURVE_POINT_OCTET.contains(&typ) {
        panic!("no support hash_to_curve type {}", typ)
    }

    Ok(array)
}

#[pyfunction]
pub(crate) fn point_octet_unmarshal(
    typ: &str,
    data: &[u8],
    count: usize,
) -> PyResult<PyArrowType<ArrayData>> {
    if !SUPPORT_CURVE_POINT_OCTET.contains(&typ) {
        panic!("no support hash_to_curve type {}", typ)
    }

    let mut build = BinaryBuilder::new();
    if count == 0 {
        return Ok(PyArrowType(build.finish().into_data()));
    }
    let block_size: usize = data.len() / count;

    for i in 0..count {
        build.append_value(&data[i * block_size..i * block_size + block_size])
    }

    Ok(PyArrowType(build.finish().into_data()))
}

pub(crate) static SUPPORT_CURVE: [&str; 1] = ["CURVE_TYPE_CURVE25519"];
pub(crate) static SUPPORT_HASHTOCURVE: [&str; 1] =
    ["HASH_TO_CURVE_STRATEGY_DIRECT_HASH_AS_POINT_X"];
pub(crate) static SUPPORT_CURVE_POINT_OCTET: [&str; 1] = ["POINT_OCTET_FORMAT_UNCOMPRESSED"];

//v2-----------------------------------------------------------------------------------------------

pub(crate) static SUPPORT_CURVE_V2: [&str; 2] = ["CURVE_TYPE_CURVE25519", "CURVE_TYPE_FOURQ"];
// pub(crate) static SUPPORT_HASHTOCURVE_V2: [&str; 2] = [
//     "HASH_TO_CURVE_STRATEGY_DIRECT_HASH_AS_POINT_X",
//     "HASH_TO_CURVE_STRATEGY_DIRECT_HASH_AS_FOURQ_POINT",
// ];
// pub(crate) static SUPPORT_CURVE_POINT_OCTET_V2: [&str; 2] =
//     ["POINT_OCTET_FORMAT_UNCOMPRESSED", "POINT_FOURQ_FORMAT"];

pub(crate) trait SecretFFI: Send {
    fn encrypt_impl(&self, data: &[u8], build: &mut BinaryBuilder);
    fn diffie_hellman_impl(&self, data: &[u8], build: &mut BinaryBuilder);
    fn new_impl(key: &[u8]) -> Box<dyn SecretFFI>
    where
        Self: Sized;
}

pub(crate) mod curve_ffi_25519 {
    use crate::curve::SecretFFI;
    use arrow::array::BinaryBuilder;
    use curve25519_dalek::montgomery::MontgomeryPoint;
    use curve25519_dalek::scalar::Scalar;

    pub(crate) struct Secret(Scalar);

    impl SecretFFI for Secret {
        fn encrypt_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
            build.append_value(
                (MontgomeryPoint(data.try_into().expect("encrypt accpet 32 bytes pubkey"))
                    * self.0)
                    .as_bytes(),
            );
        }

        fn diffie_hellman_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
            build.append_value(
                (MontgomeryPoint(
                    data.try_into()
                        .expect("diffie_hellman accpet 32 bytes pubkey"),
                ) * self.0)
                    .as_bytes(),
            );
        }

        fn new_impl(key: &[u8]) -> Box<dyn SecretFFI> {
            Box::new(Secret(Scalar::from_bytes_mod_order(
                key.try_into().unwrap(),
            )))
        }
    }
}

pub(crate) mod curve_ffi_fourq {
    use crate::curve::SecretFFI;
    use arrow::array::BinaryBuilder;
    use fourq::point::Point;
    use fourq::scalar::Scalar;

    pub(crate) struct Secret(Scalar);

    impl SecretFFI for Secret {
        fn encrypt_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
            let point = Point::from_hash(data);
            let r = point * self.0;
            let mut bytes = [0u8; 32];
            r.encode(&mut bytes);
            build.append_value(bytes);
        }

        fn diffie_hellman_impl(&self, data: &[u8], build: &mut BinaryBuilder) {
            let point = Point::decode(data);
            let r = point * self.0;
            let mut bytes = [0u8; 32];
            r.encode(&mut bytes);
            build.append_value(bytes);
        }

        fn new_impl(key: &[u8]) -> Box<dyn SecretFFI> {
            let std_key: [u8; 32] = key.try_into().unwrap();
            Box::new(Secret(std_key.try_into().unwrap()))
        }
    }
}

use curve_ffi_25519::Secret as Secret_25519;
use curve_ffi_fourq::Secret as Secret_fourq;

#[pyclass(module = "crypto.curve", name = "CurveV2")]
pub(crate) struct SecretV2(Box<dyn SecretFFI>);

impl SecretV2 {
    fn run_impl<F>(
        &self,
        func: F,
        array: PyArrowType<ArrayData>,
    ) -> PyResult<PyArrowType<ArrayData>>
    where
        F: Fn(&[u8], &mut BinaryBuilder),
    {
        let array: ArrayData = array.0;
        let array: Arc<dyn Array> = make_array(array);
        let array: &BinaryArray = array
            .as_any()
            .downcast_ref()
            .ok_or_else(|| PyValueError::new_err("expected binary array"))?;

        let mut build = BinaryBuilder::new();
        for i in array.iter() {
            func(i.unwrap(), &mut build);
        }

        Ok(PyArrowType(build.finish().into_data()))
    }
}

#[pymethods]
impl SecretV2 {
    #[new]
    #[pyo3(signature = (typ,key))]
    fn pynew(typ: &str, key: Option<[u8; 32]>) -> PyResult<Self> {
        if !SUPPORT_CURVE_V2.contains(&typ) {
            panic!("no support curve type {}", typ)
        }

        let std_key = key.unwrap_or_else(|| {
            let mut bytes: [u8; 32] = [0; 32];
            StdRng::from_entropy().fill_bytes(&mut bytes);
            bytes
        });

        match typ {
            "CURVE_TYPE_CURVE25519" => Ok(SecretV2(Secret_25519::new_impl(&std_key))),
            "CURVE_TYPE_FOURQ" => Ok(SecretV2(Secret_fourq::new_impl(&std_key))),
            &_ => Err(PyValueError::new_err("Invalid curve type")),
        }
    }

    #[pyo3(text_signature = "($self, array)")]
    fn encrypt(&self, array: PyArrowType<ArrayData>) -> PyResult<PyArrowType<ArrayData>> {
        self.run_impl(|data, build| self.0.encrypt_impl(data, build), array)
    }

    #[pyo3(text_signature = "($self, array)")]
    fn diffie_hellman(&self, array: PyArrowType<ArrayData>) -> PyResult<PyArrowType<ArrayData>> {
        self.run_impl(|data, build| self.0.diffie_hellman_impl(data, build), array)
    }
}
