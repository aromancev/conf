<template>
  <div class="logo">C<span class="o"></span>nfa</div>
</template>

<style scoped lang="sass">

$spacing: 0.05em

.logo
  color: var(--color-font)
  user-select: none
  cursor: default
  -webkit-tap-highlight-color: rgba(0,0,0,0)
  letter-spacing: $spacing

.o
  display: inline-block
  width: 0.7em
  height: 0.7em
  background-color: var(--color-red)
  border-radius: 50%
  transform: translateY(15%)
  margin-right: $spacing
</style>
