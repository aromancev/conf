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
    <circle fill="var(--color-red)" cx="8" cy="54" r="6">
      <animateTransform
        attributeName="transform"
        dur="1.2s"
        type="rotate"
        from="0 50 48"
        to="360 50 52"
        repeatCount="indefinite"
      />
    </circle>
  </svg>
</template>

<style scoped lang="sass">
svg
  width: 70px
  height: 70px
  max-width: 100%
  max-height: 100%
</style>
