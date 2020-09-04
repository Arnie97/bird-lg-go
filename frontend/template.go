package main

import (
	"text/template"
)

type tmplArguments struct {
	// Global options
	Options map[string]string
	Servers []string

	// Parameters related to current request
	AllServersLinkActive bool
	AllServersURL        string

	// Whois specific handling (for its unique URL)
	IsWhois     bool
	WhoisTarget string

	URLOption  string
	URLServer  string
	URLCommand string

	// Generated content to be displayed
	Title   string
	Brand   string
	Content string
}

var tmpl = template.Must(template.New("tmpl").Parse(`
<!DOCTYPE html>
<html lang="en-US">
<head>
<meta http-equiv="Content-Type" content="text/html;charset=UTF-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
<meta name="renderer" content="webkit">
<title>{{ .Title }}</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.5.1/dist/css/bootstrap.min.css" integrity="sha256-VoFZSlmyTXsegReQCNmbXrS4hBBUl/cexZvPmPWoJsY=" crossorigin="anonymous">
<meta name="robots" content="noindex, nofollow">
<style>
	.container h2 {
		font-size: 1.5rem;
		margin: 48px 0px 20px;
	}
	.nav-link.active{
		font-weight: bold;
	}
</style>
</head>
<body>

<nav class="navbar navbar-expand-lg navbar-light bg-light fixed-top border-bottom">
	<a class="navbar-brand" href="/">{{ .Brand }}</a>
	<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
		<span class="navbar-toggler-icon"></span>
	</button>

	<div class="collapse navbar-collapse" id="navbarSupportedContent">
		<ul class="navbar-nav mr-auto">
			<li class="nav-item">
				<a class="nav-link{{ if .AllServersLinkActive }} active{{ end }}" href="/{{ .URLOption }}/{{ .AllServersURL }}/{{ .URLCommand }}">all</a>
			</li>
			{{ range $k, $v := .Servers }}
			<li class="nav-item">
				<a class="nav-link{{ if eq $.URLServer $v }} active{{ end }}" href="/{{ $.URLOption }}/{{ $v }}/{{ $.URLCommand }}">{{ $v }}</a>
			</li>
			{{ end }}
		</ul>
		{{ $option := .URLOption }}
		{{ $target := .URLCommand }}
		{{ if .IsWhois }}
			{{ $option = "whois" }}
			{{ $target = .WhoisTarget }}
		{{ end }}
		<form class="form-inline" action="/redir" method="GET">
			<div class="input-group">
				<select name="action" class="form-control">
					{{ range $k, $v := .Options }}
					<option value="{{ $k }}"{{ if eq $k $option }} selected{{end}}>{{ $v }}</option>
					{{ end }}
				</select>
				<input name="server" class="d-none" value="{{ .URLServer }}">
				<input name="target" class="form-control" placeholder="Target" aria-label="Target" value="{{ $target }}">
				<div class="input-group-append">
					<button class="btn btn-outline-success" type="submit">&raquo;</button>
				</div>
			</div>
		</form>
	</div>
</nav>

<div class="container px-4 py-5">
	{{ .Content }}
</div>

<script src="https://cdn.jsdelivr.net/npm/jquery@3.5.1/dist/jquery.min.js" integrity="sha256-9/aliU8dGd2tb6OSsuzixeV4y/faTqgFtohetphbbj0=" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@4.5.1/dist/js/bootstrap.min.js" integrity="sha256-0IiaoZCI++9oAAvmCb5Y0r93XkuhvJpRalZLffQXLok=" crossorigin="anonymous"></script>
</body>
</html>
`))

const peeringForm = `
<form>
<div class="form-group row">
	<label for="aliceASN" class="col-xs-12 col-md-4 col-lg-3">Our AS Number</label>
	<input id="aliceASN" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="aliceLoc" class="col-xs-12 col-md-4 col-lg-3">Location (IATA Identifier)</label>
	<input id="aliceLoc" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="ourPublicIPv4" class="col-xs-12 col-md-4 col-lg-3">Public IPv4 Address</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="ourPublicIPv4" value="47.240.24.53" readonly>
</div>
<div class="form-group row">
	<label for="aliceWG" class="col-xs-12 col-md-4 col-lg-3">Public IPv6 Address</label>
	<input id="aliceWG" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="aliceIPv4" class="col-xs-12 col-md-4 col-lg-3">DN42 IPv4 Address</label>
	<input id="aliceIPv4" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="aliceLink" class="col-xs-12 col-md-4 col-lg-3">Link Local IPv6 Address</label>
	<input id="aliceLink" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="alicePubl" class="col-xs-12 col-md-4 col-lg-3">WireGuard Public Key</label>
	<input id="alicePubl" type="text" class="form-control col-xs-12 col-md-8 col-lg-6" readonly>
</div>
<div class="form-group row">
	<label for="ourWGPort" class="col-xs-12 col-md-4 col-lg-3">WireGuard Listen Port</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="ourWGPort" placeholder="(Last 5 digits of peer ASN)" value="" readonly>
</div>
</form>

<h2>your endpoint</h2>

<form>
<div class="form-group row">
	<label for="bobAS" class="col-xs-12 col-md-4 col-lg-3">Peer AS Number</label>
	<div class="input-group col-xs-12 col-md-8 col-lg-6 p-0">
		<div class="input-group-prepend"><span class="input-group-text">AS</span></div>
		<input type="text" class="form-control" id="peerAS" name="peerAS" placeholder="424242xxxx">
	</div>
</div>
<div class="form-group row">
	<label for="bobLoc" class="col-xs-12 col-md-4 col-lg-3">Location (IATA Identifier)</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerLocation" name="peerLocation" placeholder="XXX">
</div>
<div class="form-group row">
	<label for="peerPublicIPv4" class="col-xs-12 col-md-4 col-lg-3">Public IPv4 Address</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerPublicIPv4" name="peerPublicIPv6" placeholder="xxx.xxx.xxx.xxx">
</div>
<div class="form-group row">
	<label for="peerPublicIPv6" class="col-xs-12 col-md-4 col-lg-3">Public IPv6 Address</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerPublicIPv6" name="peerPublicIPv6" placeholder="2xxx:xxxx:xxxx:c0de::cafe">
</div>
<div class="form-group row">
	<label for="bobIPv4" class="col-xs-12 col-md-4 col-lg-3">DN42 IPv4 Address</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerDN42IPv4" name="peerDN42IPv4" placeholder="172.2x.xxx.xxx">
</div>
<div class="form-group row">
	<label for="bobLink" class="col-xs-12 col-md-4 col-lg-3">Link Local IPv6 Address</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerLinkLocalIPv6" name="peerLinkLocalIPv6" placeholder="fe80::xxxx">
</div>
<div class="form-group row">
	<label for="bobPubl" class="col-xs-12 col-md-4 col-lg-3">WireGuard Public Key</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerWGPublicKey" name="peerWGPublicKey" required>
</div>
<div class="form-group row">
	<label for="peerWGPort" class="col-xs-12 col-md-4 col-lg-3">WireGuard Listen Port</label>
	<input type="text" class="form-control col-xs-12 col-md-8 col-lg-6" id="peerWGPort" name="peerWGPort" placeholder="20977" value="20977" required>
</div>

<h2>bgp preferences</h2>
<div class="form-group row">
	<label for="multiProtocol" class="col-xs-12 col-md-4 col-lg-3">Multi-protocol Session</label>
	<select class="form-control col-xs-12 col-md-8 col-lg-6" id="multiProtocol" name="multiProtocol">
		<option value="mp">Multi-protocol BGP over IPv6 transport (Preferred)</option>
		<option value="dual">IPv4 routes over IPv4 transport, and IPv6 routes over IPv6</option>
		<option value="v6">Route only IPv6 prefixes (over IPv6 transport)</option>
		<option value="v4">Route only IPv4 prefixes (over IPv4 transport)</option>
	</select>
</div>
<div class="form-group row">
	<label for="linkLatency" class="col-xs-12 col-md-4 col-lg-3">Link Latency</label>
	<select class="form-control col-xs-12 col-md-8 col-lg-6" id="linkLatency" name="linkLatency">
		<option value="1">&le; 2.7ms (64511, 1)</option>
		<option value="2">&le; 7.3ms (64511, 2)</option>
		<option value="3" selected>&le; 20ms (64511, 3)</option>
		<option value="4">&le; 55ms (64511, 4)</option>
		<option value="5">&le; 148ms (64511, 5)</option>
		<option value="6">&le; 403ms (64511, 6)</option>
		<option value="7">&le; 1097ms (64511, 7)</option>
		<option value="8">&le; 2981ms (64511, 8)</option>
		<option value="9">&gt; 2981ms (64511, 9)</option>
	</select>
</div>
<div class="form-group row">
	<label for="linkBandwidth" class="col-xs-12 col-md-4 col-lg-3">Link Bandwidth</label>
	<select class="form-control col-xs-12 col-md-8 col-lg-6" id="linkBandwidth" name="linkBandwidth">
		<option value="20">&lt; 100kbps (64511, 20)</option>
		<option value="21">&ge; 100kbps (64511, 21)</option>
		<option value="22">&ge; 1Mbps (64511, 22)</option>
		<option value="23">&ge; 10Mbps (64511, 23)</option>
		<option value="24" selected>&ge; 100Mbps (64511, 24)</option>
		<option value="25">&ge; 1Gbps (64511, 25)</option>
		<option value="26">&ge; 10Gbps (64511, 26)</option>
	</select>
</div>
<button type="submit" class="btn btn-primary">Submit</button>
</form>

<script>
document.addEventListener('DOMContentLoaded', function() {
	document.querySelector('#peerAS').addEventListener('change', function(event) {
		var peerAS = event.target.value;
		document.querySelector('#ourWGPort').value = peerAS.slice(peerAS.length > 5? peerAS.length - 5: 0);
		document.querySelector('#peerLinkLocalIPv6').value = 'fe80::' + peerAS.slice(peerAS.length > 4? peerAS.length - 4: 0);
	});

	var roles = ['Alice', 'Bob'];
	var fields = ['ASN', 'Loc', 'IPv4', 'Link', 'Publ'];
	for (var i = 0; i < roles.length; i++) {
		for (var j = 0; j < fields.length; j++) {
			var src = info[roles[i]][fields[j].toLowerCase()],
				dest = document.querySelector('#' + roles[i].toLowerCase() + fields[j]);
			if (src && dest) {
				dest.value = src;
			}
		}
	}
});
</script>
`
