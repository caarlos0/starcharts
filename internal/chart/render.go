package chart

import (
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
)

func (c *Chart) Render(w io.Writer) {

	path := svg.Path().
		MoveTo(38, 351).
		LineTo(48, 348).
		LineTo(57, 345).
		LineTo(67, 343).
		LineTo(76, 340).
		LineTo(85, 335).
		LineTo(95, 333).
		LineTo(104, 330).
		LineTo(113, 325).
		LineTo(123, 325).
		LineTo(132, 323).
		LineTo(142, 318).
		LineTo(151, 317).
		LineTo(160, 309).
		LineTo(170, 313).
		//Attr("d", " L 179 306 L 188 303 L 198 300 L 207 294 L 216 298 L 226 297 L 235 289 L 245 286 L 254 283 L 263 283 L 273 281 L 282 269 L 291 264 L 301 265 L 310 260 L 319 269 L 329 268 L 338 255 L 348 245 L 357 260 L 366 245 L 376 237 L 385 247 L 394 245 L 404 243 L 413 227 L 422 220 L 432 223 L 441 237 L 451 227 L 460 228 L 469 223 L 479 225 L 488 201 L 497 197 L 507 193 L 516 193 L 525 198 L 535 209 L 544 183 L 554 185 L 563 196 L 572 182 L 582 175 L 591 171 L 600 164 L 610 184 L 619 156 L 628 183 L 638 148 L 647 168 L 657 142 L 666 162 L 675 165 L 685 157 L 694 124 L 703 152 L 713 130 L 722 153 L 731 116 L 741 120 L 750 139 L 760 100 L 769 98 L 778 122 L 788 113 L 797 109 L 806 87 L 816 105 L 825 95 L 834 86 L 844 84 L 853 116 L 863 63 L 872 67 L 881 86 L 891 97 L 900 63 L 909 50 L 919 49 L 928 79 L 937 89 L 947 49 L 956 36 L 965 54").
		Attr("fill", "none").
		Attr("stroke", "black").
		Attr("stroke-width", "2")

	svgElement := svg.SVG().
		Attr("width", svg.Px(DefaultChartWidth)).
		Attr("height", svg.Px(DefaultChartHeight)).
		ContentFunc(func() string {
			return svg.Style().
				Attr("type", "text/css").
				Content(`
					path {
						stroke: red;
						stroke-width: 2;
						fill: none;
					}
				`).
				String()
		}).
		Content(path.String())

	svgElement.WriteTo(w)
}
