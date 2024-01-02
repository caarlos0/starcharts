package chart

const LightStyles = `
path { fill: none; stroke: rgb(51,51,51); }
path.series { stroke: #6b63ff; }
rect.background { fill: rgb(255,255,255); stroke: none; }

text {
	stroke-width: 0;
	stroke: none;
	fill: rgba(51,51,51,1.0);
	font-size: 12.8px;
	font-family: 'Roboto Medium', sans-serif;
}
`

const DarkStyles = `
path { fill: none; stroke: rgb(51,51,51); }
path.series { stroke: #6b63ff; }
rect.background { fill: rgb(255,255,255); stroke: none; }

text {
	stroke-width: 0;
	stroke: none;
	fill: rgba(51,51,51,1.0);
	font-size: 12.8px;
	font-family: 'Roboto Medium', sans-serif;
}

path { stroke: rgb(230, 237, 243); }
path.series { stroke: #6b63ff; }
text { fill: rgb(230, 237, 243); }
rect.background { fill: rgb(0,0,0); }
`

const AdaptiveStyles = `
path { fill: none; stroke: rgb(51,51,51); }
path.series { stroke: #6b63ff; }
rect.background { fill: none; stroke: none; }

text {
	stroke-width: 0;
	stroke: none;
	fill: rgba(51,51,51,1.0);
	font-size: 12.8px;
	font-family: 'Roboto Medium', sans-serif;
}

@media (prefers-color-scheme: dark) {
	path { stroke: rgb(230, 237, 243); }
	path.series { stroke: #6b63ff; }
	text { fill: rgb(230, 237, 243); }
}
`
