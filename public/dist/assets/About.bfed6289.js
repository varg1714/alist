import{d as a,X as r,f as t,Z as n}from"./index.7e2ddf5f.js";import{o}from"./index.0b63cd90.js";import{M as s}from"./Markdown.d2f633b7.js";const i=async()=>await(await fetch("https://jsd.nn.ci/gh/alist-org/alist@main/README.md")).text(),u=()=>{a(),o("manage.sidemenu.about");const[e]=r(i);return t(n,{get loading(){return e.loading},get children(){return t(s,{get children(){return e()}})}})};export{u as default};
