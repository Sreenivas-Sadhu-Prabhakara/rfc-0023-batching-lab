package main

import "html/template"

var page = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>RFC 23 · Control message batching</title>
<style>
:root { background: #eef4f3; color: #132e34; font: 17px/1.55 system-ui, sans-serif; }
main { max-width: 900px; margin: auto; padding: 36px 22px; }
h1 { font-size: clamp(2rem, 6vw, 3.4rem); line-height: 1.08; margin: .25em 0; }
h2 { margin: 1.15em 0 .35em; }
a { color: #115e67; }
.table { overflow-x: auto; background: white; border: 1px solid #a5c0bb; border-radius: 10px; }
table { text-align: left; width: 100%; border-collapse: collapse; }
th, td { padding: 7px 12px; border-top: 1px solid #d5e4e1; vertical-align: top; }
thead { background: #dbece8; }
</style></head>
<body><main>
<h1>RFC 23 · Control message batching</h1>
<p>{{.Commands}} commands, {{.AckDelay}} ticks until the link is ready again, and a {{.MaxBytes}}-byte payload limit. Both policies send as soon as the link is ready.</p>
<div class="table" role="region" aria-label="Policy comparison" tabindex="0"><table>
<thead><tr><th scope="col">Policy</th><th scope="col">Sends</th><th scope="col">Final ready tick</th><th scope="col">Longest queue wait</th></tr></thead>
<tbody>{{range .Runs}}<tr><th scope="row">{{.Policy}}</th><td>{{len .Sends}}</td><td>{{.LastReady}}</td><td>{{.MaxQueueWait}} ticks</td></tr>{{end}}</tbody>
</table></div>
{{range .Runs}}
<h2>{{.Policy}}</h2>
<div class="table" role="region" aria-label="{{.Policy}} transmissions" tabindex="0"><table>
<thead><tr><th scope="col">Send tick</th><th scope="col">Ready tick</th><th scope="col">Payload hex</th><th scope="col">Commands decoded in order</th></tr></thead>
<tbody>{{range .Sends}}<tr><th scope="row">{{.Send}}</th><td>{{.Ready}}</td><td><code>{{.Hex}}</code></td><td>{{range $i, $command := .Commands}}{{if $i}} · {{end}}{{$command.Name}}{{end}}</td></tr>{{end}}</tbody>
</table></div>{{end}}
<p>The local Go simulator generated this saved report. <a href="https://www.rfc-editor.org/info/rfc23/">Official RFC 23</a> · <a href="https://github.com/Sreenivas-Sadhu-Prabhakara/rfc-0023-batching-lab">Source and run instructions</a></p>
</main></body></html>
`))
