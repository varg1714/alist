import{f as n,Z as d,o,bG as i}from"./index.3efbc392.js";import{d as s}from"./useUtil.d7d390bc.js";import{M as m}from"./Markdown.d2f3b361.js";import"./api.0bcae678.js";const p=()=>{const[t]=s(),a=e=>o.obj.name.endsWith(".md")?e:"```"+i(o.obj.name)+`
`+e+"\n```";return n(d,{get loading(){return t.loading},get children(){return n(m,{get children(){var e,r;return a((r=(e=t())==null?void 0:e.content)!=null?r:"")}})}})};export{p as default};
