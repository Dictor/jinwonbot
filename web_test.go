package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	openDocument = `<!DOCTYPE html>
<html lang="en-US">
  <head>
    <meta charset="UTF-8">

<!-- Begin Jekyll SEO tag v2.6.1 -->
<title>바라미실은 열렸는가? | 바라미</title>
<meta name="generator" content="Jekyll v3.9.0" />
<meta property="og:title" content="바라미실은 열렸는가?" />
<meta property="og:locale" content="en_US" />
<link rel="canonical" href="https://ibarami.github.io/" />
<meta property="og:url" content="https://ibarami.github.io/" />
<meta property="og:site_name" content="바라미" />
<script type="application/ld+json">
{"@type":"WebSite","url":"https://ibarami.github.io/","headline":"바라미실은 열렸는가?","name":"바라미","@context":"https://schema.org"}</script>
<!-- End Jekyll SEO tag -->

    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="theme-color" content="#157878">
    <link rel="stylesheet" href="/assets/css/style.css?v=b02226cfc4f25423b737d180ec90ada02a28b5a9">
  </head>
  <body>
    <section class="page-header">
      <h1 class="project-name">바라미</h1>
      <h2 class="project-tagline"></h2>


    </section>

    <section class="main-content">
      <h2 id="바라미실은-열렸는가">바라미실은 열렸는가?</h2>

<p>현재 바라미실 상태: <strong>열림</strong></p>

<h3 id="바라미란">바라미란</h3>

<div class="language-markdown highlighter-rouge"><div class="highlight"><pre class="highlight"><code>한양대학교 전자전기컴퓨터학술 동아리입니다.

바라미는 친목과 작품활동을 즐길 수 있는 동아리입니다.
</code></pre></div></div>

<h3 id="contact">Contact</h3>

<p>저희 동아리에 관심을 가지시는 분이 계시다면 ibarami.com으로 들어와주세요!</p>

<h3 id="참고사항">참고사항</h3>

<p>github.io의 반영속도로 인해서 상태가 바뀌는데 1분 내외로 소요됩니다.</p>

<h3 id="만든이">만든이</h3>

<p>한양대학교 바라미 24기, 융합전자공학과 20학번 주진원</p>

      <footer class="site-footer">

        <span class="site-footer-credits">This page was generated by <a href="https://pages.github.com">GitHub Pages</a>.</span>
      </footer>
    </section>


  </body>
</html>`
	closeDocument = `<!DOCTYPE html>
	<html lang="en-US">
	  <head>
		<meta charset="UTF-8">
	
	<!-- Begin Jekyll SEO tag v2.6.1 -->
	<title>바라미실은 열렸는가? | 바라미</title>
	<meta name="generator" content="Jekyll v3.9.0" />
	<meta property="og:title" content="바라미실은 열렸는가?" />
	<meta property="og:locale" content="en_US" />
	<link rel="canonical" href="https://ibarami.github.io/" />
	<meta property="og:url" content="https://ibarami.github.io/" />
	<meta property="og:site_name" content="바라미" />
	<script type="application/ld+json">
	{"@type":"WebSite","url":"https://ibarami.github.io/","headline":"바라미실은 열렸는가?","name":"바라미","@context":"https://schema.org"}</script>
	<!-- End Jekyll SEO tag -->
	
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<meta name="theme-color" content="#157878">
		<link rel="stylesheet" href="/assets/css/style.css?v=b02226cfc4f25423b737d180ec90ada02a28b5a9">
	  </head>
	  <body>
		<section class="page-header">
		  <h1 class="project-name">바라미</h1>
		  <h2 class="project-tagline"></h2>
	
	
		</section>
	
		<section class="main-content">
		  <h2 id="바라미실은-열렸는가">바라미실은 열렸는가?</h2>
	
	<p>현재 바라미실 상태: <strong>닫힘</strong></p>
	
	<h3 id="바라미란">바라미란</h3>
	
	<div class="language-markdown highlighter-rouge"><div class="highlight"><pre class="highlight"><code>한양대학교 전자전기컴퓨터학술 동아리입니다.
	
	바라미는 친목과 작품활동을 즐길 수 있는 동아리입니다.
	</code></pre></div></div>
	
	<h3 id="contact">Contact</h3>
	
	<p>저희 동아리에 관심을 가지시는 분이 계시다면 ibarami.com으로 들어와주세요!</p>
	
	<h3 id="참고사항">참고사항</h3>
	
	<p>github.io의 반영속도로 인해서 상태가 바뀌는데 1분 내외로 소요됩니다.</p>
	
	<h3 id="만든이">만든이</h3>
	
	<p>한양대학교 바라미 24기, 융합전자공학과 20학번 주진원</p>
	
		  <footer class="site-footer">
	
			<span class="site-footer-credits">This page was generated by <a href="https://pages.github.com">GitHub Pages</a>.</span>
		  </footer>
		</section>
	
	
	  </body>
	</html>`
)

func TestWebRequesting(t *testing.T) {
	assert.True(t, isDoorOpen(openDocument))
	assert.False(t, isDoorOpen(closeDocument))
}
