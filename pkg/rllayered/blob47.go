package rllayered

// 47-tile "blob" autotile, used by the GMS2 "Auto Tile (47 brush)" template
// and similar tilesets. The mask is 8-directional:
//
//   bit 0 = N         bit 4 = NE
//   bit 1 = S         bit 5 = NW
//   bit 2 = W         bit 6 = SE
//   bit 3 = E         bit 7 = SW
//
// Corner pruning: a diagonal bit only counts when BOTH of its adjacent
// cardinals are also set (an inward corner only "exists" when the wall wraps
// the full corner). After pruning there are exactly 47 unique masks.
//
// The engine assigns each pruned mask a stable index 0..46 by sorting the 47
// valid masks numerically. The tile definition holds 47 variants; whichever
// variant has `Variant == N` provides the sprite for index N. To match the
// GMS2 8×6 layout, point variant N at the (col N%8, row N/8) cell of your
// tilesheet (skipping the canonical blank slot — see your generator's docs).

const (
	BlobBitN  = 1 << 0
	BlobBitS  = 1 << 1
	BlobBitW  = 1 << 2
	BlobBitE  = 1 << 3
	BlobBitNE = 1 << 4
	BlobBitNW = 1 << 5
	BlobBitSE = 1 << 6
	BlobBitSW = 1 << 7
)

// PruneBlobMask zeroes any diagonal bit whose adjacent cardinals aren't both
// set, leaving a canonical mask that uniquely identifies the visual blob.
func PruneBlobMask(m uint8) uint8 {
	if m&(BlobBitN|BlobBitE) != BlobBitN|BlobBitE {
		m &^= BlobBitNE
	}
	if m&(BlobBitN|BlobBitW) != BlobBitN|BlobBitW {
		m &^= BlobBitNW
	}
	if m&(BlobBitS|BlobBitE) != BlobBitS|BlobBitE {
		m &^= BlobBitSE
	}
	if m&(BlobBitS|BlobBitW) != BlobBitS|BlobBitW {
		m &^= BlobBitSW
	}
	return m
}

// blob47Index maps a pruned mask → variant index 0..46. Built once at init.
var blob47Index [256]int8

// Blob47Count is the number of unique pruned masks (always 47).
const Blob47Count = 47

func init() {
	for i := range blob47Index {
		blob47Index[i] = -1
	}
	// Enumerate every raw mask, prune, and assign indices in ascending
	// pruned-mask order. This produces a stable 0..46 mapping that doesn't
	// depend on any external template's row-major order.
	seen := map[uint8]bool{}
	uniques := make([]uint8, 0, Blob47Count)
	for raw := 0; raw < 256; raw++ {
		p := PruneBlobMask(uint8(raw))
		if !seen[p] {
			seen[p] = true
			uniques = append(uniques, p)
		}
	}
	// uniques is already in ascending order because we iterate raw 0..255
	// and the smallest pruning of each new mask appears first.
	for i, m := range uniques {
		blob47Index[m] = int8(i)
	}
}

// Blob47Slot returns the variant index 0..46 for an 8-direction neighbor mask
// (pre or post pruning).
func Blob47Slot(rawMask uint8) int {
	idx := blob47Index[PruneBlobMask(rawMask)]
	if idx < 0 {
		return 0
	}
	return int(idx)
}
