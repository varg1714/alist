import{d as a,X as r,f as t,Z as n}from"./index.c7be97ab.js";import{o}from"./index.f3072a2b.js";import{M as s}from"./Markdown.b592c64e.js";const i=async()=>await(await fetch("https://jsd.nn.ci/gh/alist-org/alist@main/README.md")).text(),u=()=>{a(),o("manage.sidemenu.about");const[e]=r(i);return t(n,{get loading(){return e.loading},get children(){return t(s,{get children(){return e()}})}})};export{u as default};
