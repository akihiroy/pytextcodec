#!/usr/bin/env python3
"""
Script to generate CP932 encoder and decoder test data

- Encoder data: key is unicode codepoint, value is cp932 byte sequence
- Decoder data: key is cp932 byte sequence, value is unicode codepoint
"""

import os
from typing import Dict

import msgpack


def generate_cp932_encoder_data() -> Dict[str, bytes]:
    """
    Generate test data for CP932 encoder
    key: unicode codepoint (0~0x10FFFF) as string
    value: cp932 byte sequence
    """
    encoder_data = {}

    # Process entire Unicode range (0x000000 - 0x10FFFF)
    for i in range(0x110000):
        try:
            char = chr(i)
            encoded = char.encode("cp932")
            if encoded:
                encoder_data[f"U+{i:06X}"] = encoded
        except (UnicodeEncodeError, LookupError):
            pass

    return encoder_data


def generate_cp932_decoder_data() -> Dict[str, str]:
    """
    Generate test data for CP932 decoder
    key: cp932 byte sequence as hex string
    value: unicode codepoint as string
    """
    decoder_data = {}

    for c0 in range(0x100):  # 0x00-0xFF
        for c1 in range(0x100):  # 0x00-0xFF
            # If c0 corresponds to single-byte character range, perform single-byte decoding
            if c0 <= 0x80 or (0xA0 <= c0 <= 0xDF) or (0xFD <= c0 <= 0xFF):
                try:
                    byte_seq = bytes([c0])
                    decoded = byte_seq.decode("cp932")
                    if decoded:
                        decoder_data[f"{c0:02X}"] = f"U+{ord(decoded):06X}"
                except (UnicodeDecodeError, LookupError):
                    pass
                break  # Skip 2-byte processing for single-byte characters

            # 2-byte character processing
            try:
                byte_seq = bytes([c0, c1])
                decoded = byte_seq.decode("cp932")
                if decoded:
                    decoder_data[f"{c0:02X}{c1:02X}"] = f"U+{ord(decoded):06X}"
            except (UnicodeDecodeError, LookupError):
                pass

    return decoder_data


def save_testdata(encoder_data: Dict[str, bytes], decoder_data: Dict[str, str]):
    """Save test data to MessagePack files"""
    output_dir = os.path.dirname(__file__)

    # Save encoder data
    encoder_file = os.path.join(output_dir, "cp932_encoder.msgpack")
    with open(encoder_file, "wb") as f:
        msgpack.pack(encoder_data, f)
    print(f"Encoder test data saved: {encoder_file}")
    print(f"  Data count: {len(encoder_data)}")

    # Save decoder data
    decoder_file = os.path.join(output_dir, "cp932_decoder.msgpack")
    with open(decoder_file, "wb") as f:
        msgpack.pack(decoder_data, f)
    print(f"Decoder test data saved: {decoder_file}")
    print(f"  Data count: {len(decoder_data)}")


def main():
    """Main processing"""
    print("Starting japanese test data generation...")

    print("Generating encoder data for CP932...")
    encoder_data = generate_cp932_encoder_data()

    print("Generating decoder data for CP932...")
    decoder_data = generate_cp932_decoder_data()

    save_testdata(encoder_data, decoder_data)

    print("Test data generation completed!")


if __name__ == "__main__":
    main()
