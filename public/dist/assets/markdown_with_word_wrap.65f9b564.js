import{f as o,Z as s,o as n,bG as d}from"./index.9265efac.js";import{d as i}from"./useUtil.7842b7d6.js";import{M as m}from"./Markdown.377e7797.js";import"./api.bc5114a2.js";const g=()=>{const[r]=i(),a=e=>n.obj.name.endsWith(".md")?e:"```"+d(n.obj.name)+`
`+e+"\n```";return o(s,{get loading(){return r.loading},get children(){return o(m,{class:"word-wrap",get children(){var e,t;return a((t=(e=r())==null?void 0:e.content)!=null?t:"")}})}})};export{g as default};
