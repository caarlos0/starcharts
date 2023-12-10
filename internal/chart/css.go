package chart

const chartCss = `
path { fill: none; stroke: rgb(51,51,51); }
path.series { stroke: rgb(129,199,239); }
rect.background { fill: rgb(255,255,255); stroke: none; }

text {
	stroke-width: 0;
	stroke: none;
	fill: rgba(51,51,51,1.0);
	font-size: 12.8px;
	font-family: 'Roboto Medium', sans-serif;
}

@media (prefers-color-scheme: dark) {
	path { stroke: rgb(230, 237, 243); }
	path.series { stroke: rgb(47, 129, 247); }
	text { fill: rgb(230, 237, 243); }
	rect.background { fill: rgb(0,0,0);
}
`
