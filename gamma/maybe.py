

__all__ = (
    "Maybe",
    "nothing",
    "Just"
)


class Maybe():
    def __xor__(self, action): # equivalent to Haskell's >>=
        if self.__class__ == _Maybe__Nothing:
            return Nothing
        elif self.__class__ == Just:
            return action(self.value)
    
    def __bool__(self):
        return self.__class__ != _Maybe__Nothing
    

class _Maybe__Nothing(Maybe):
    def __repr__(self):
        return "Nothing"

nothing = _Maybe__Nothing()


class Just(Maybe):
    def __init__(self, v):
        self.value = v
    def __repr__(self):
        return "Just(%r)" % self.value
