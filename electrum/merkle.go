package electrum

import (
	"crypto/sha256"
)

func DoubleSha256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

func ReverseBytes(b []byte) []byte {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-1-i] = b[len(b)-1-i], b[i]
	}
	return b
}

func GetMerkleRootFromPath(txid []byte, txIndex uint64, path [][32]byte, reverse bool) []byte {
	if reverse {
		//txid = ReverseBytes(txid)
	}
	txHash := DoubleSha256(txid)
	for _, p := range path {
		var data []byte
		if txIndex%2 == 0 {
			data = append(txHash[:], p[:]...)
		} else {
			data = append(p[:], txHash[:]...)
		}
		data = DoubleSha256(data)
		copy(txHash[:], data)

		txIndex /= 2
	}
	if reverse {
		return ReverseBytes(txHash[:])
	}
	return txHash[:]
}

// fn create_merkle_branch_and_root(
//     mut hashes: Vec<Sha256dHash>,
//     mut index: usize,
// ) -> (Vec<Sha256dHash>, Sha256dHash) {
//     let mut merkle = vec![];
//     while hashes.len() > 1 {
//         if hashes.len() % 2 != 0 {
//             let last = *hashes.last().unwrap();
//             hashes.push(last);
//         }
//         index = if index % 2 == 0 { index + 1 } else { index - 1 };
//         merkle.push(hashes[index]);
//         index /= 2;
//         hashes = hashes
//             .chunks(2)
//             .map(|pair| merklize(pair[0], pair[1]))
//             .collect()
//     }
//     (merkle, hashes[0])
// }
