<template>
  <svg
    id="L3"
    class="loader"
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    x="0px"
    y="0px"
    viewBox="0 0 100 100"
    enable-background="new 0 0 0 0"
    xml:space="preserve"
  >
    <circle fill="none" stroke="var(--color-highlight-background)" stroke-width="4" cx="50" cy="50" r="44" />
    <circle fill="var(--color-red)" cx="48" cy="94" r="6">
    </circle>
  </svg>
</template>

<style scoped lang="sass">

@keyframes spin
  0%
    transform: rotate(0deg)
  
  100%
    transform: rotate(360deg)

svg
  width: 70px
  height: 70px
  max-width: 100%
  max-height: 100%
  animation: spin 5s cubic-bezier(0, 1, 1, 1) infinite
</style>
